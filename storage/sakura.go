package storage

import (
	"context"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	v4 "github.com/aws/aws-sdk-go-v2/aws/signer/v4"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/newmo-oss/ergo"
)

// SakuraObjectStorage はさくらのクラウド オブジェクトストレージ（S3互換）を使用する実装です。
type SakuraObjectStorage struct {
	client        *s3.Client
	presignClient *s3.PresignClient
	uploader      *manager.Uploader
	bucket        string
}

// SakuraStorageConfig は初期化に必要な設定情報です。
type SakuraStorageConfig struct {
	AccessKey string
	SecretKey string
	Endpoint  string // 例: "https://s3.isk01.sakurastorage.jp"
	Bucket    string
	Region    string // 基本的に "jp-north-1" (指定がない場合は自動設定しません)
}

// NewSakuraObjectStorage は新しいSakuraObjectStorageインスタンスを作成します。
func NewSakuraObjectStorage(ctx context.Context, cfg SakuraStorageConfig) (*SakuraObjectStorage, error) {
	if cfg.Bucket == "" {
		return nil, errors.New("bucket name is required")
	}

	// モックやテスト時に困らないよう、Regionが空ならデフォルトを入れるなどの配慮
	region := cfg.Region
	if region == "" {
		region = "jp-north-1" // さくらのデフォルトと仮定
	}

	// カスタムリゾルバーの定義 (v2 SDKの作法)
	resolver := aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
		return aws.Endpoint{
			URL:           cfg.Endpoint,
			SigningRegion: region,
		}, nil
	})

	// AWS SDKの設定ロード
	awsCfg, err := config.LoadDefaultConfig(ctx,
		config.WithRegion(region),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(cfg.AccessKey, cfg.SecretKey, "")),
		config.WithEndpointResolverWithOptions(resolver),
	)
	if err != nil {
		return nil, ergo.Wrap(err, "failed to load aws config")
	}

	// S3クライアントの作成
	// ForcePathStyle=true は互換ストレージで必須な場合が多い
	s3Client := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		o.UsePathStyle = true
	})

	// Presignクライアントの作成
	presignClient := s3.NewPresignClient(s3Client)

	// Uploaderの作成
	uploader := manager.NewUploader(s3Client)

	return &SakuraObjectStorage{
		client:        s3Client,
		presignClient: presignClient,
		uploader:      uploader,
		bucket:        cfg.Bucket,
	}, nil
}

func (s *SakuraObjectStorage) Upload(ctx context.Context, path string, data io.Reader) error {
	// feature/s3/manager を使用して、大きなファイルも効率的にアップロード
	_, err := s.uploader.Upload(ctx, &s3.PutObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(path),
		Body:   data,
	})
	if err != nil {
		return ergo.Wrap(err, "failed to upload to sakura object storage")
	}
	return nil
}

func (s *SakuraObjectStorage) Download(ctx context.Context, path string) (io.ReadCloser, error) {
	out, err := s.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(path),
	})
	if err != nil {
		return nil, ergo.Wrap(err, "failed to download from sakura object storage")
	}
	return out.Body, nil
}

func (s *SakuraObjectStorage) Delete(ctx context.Context, path string) error {
	_, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(path),
	})
	if err != nil {
		return ergo.Wrap(err, "failed to delete object")
	}
	return nil
}

func (s *SakuraObjectStorage) GetSignedURL(ctx context.Context, path string, method string, expires time.Duration) (string, error) {
	var req *v4.PresignedHTTPRequest
	var err error

	// methodに応じたPresignメソッドを呼び分ける
	// v2 SDKのPresignerはメソッドごとにAPIが分かれている
	switch method {
	case "GET":
		req, err = s.presignClient.PresignGetObject(ctx, &s3.GetObjectInput{
			Bucket: aws.String(s.bucket),
			Key:    aws.String(path),
		}, s3.WithPresignExpires(expires))
	case "PUT":
		req, err = s.presignClient.PresignPutObject(ctx, &s3.PutObjectInput{
			Bucket: aws.String(s.bucket),
			Key:    aws.String(path),
		}, s3.WithPresignExpires(expires))
	case "DELETE":
		req, err = s.presignClient.PresignDeleteObject(ctx, &s3.DeleteObjectInput{
			Bucket: aws.String(s.bucket),
			Key:    aws.String(path),
		}, s3.WithPresignExpires(expires))
	default:
		return "", fmt.Errorf("unsupported method for signed url: %s", method)
	}

	if err != nil {
		return "", ergo.Wrap(err, "failed to generate signed url")
	}

	return req.URL, nil
}
