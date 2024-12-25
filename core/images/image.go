package images

import (
	"fmt"
	"image"
	_ "image/gif"
	"image/jpeg"
	"image/png"
	"os"

	"github.com/nfnt/resize"
)

// ResizeTarget はリサイズのターゲットを表します。
type ResizeTarget int

const (
	// WIDTH は幅をターゲットにします。
	WIDTH ResizeTarget = iota
	// HEIGHT は高さをターゲットにします。
	HEIGHT
	// RATIO は比率をターゲットにします。
	RATIO
)

// Image は画像のリサイズ情報を保持します。
type Image struct {
	FilePath     string
	ChangeRatio  float32
	ChangeWidth  int
	ChangeHeight int
	Target       ResizeTarget
}

// ResizeToData は画像をリサイズし、バイトデータとして返却します。
// 戻り値: リサイズされた画像のバイトデータとエラー情報
func (i Image) ResizeToData() ([]byte, error) {
	filePath, err := i.ResizeToFile()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	defer func() {
		err = os.Remove(filePath)
		if err != nil {
			fmt.Printf("failed to remove file: %v\n", err)
		}
	}()

	return data, nil

}

// ResizeToFile は画像をリサイズし、ファイルとして保存します。
// 戻り値: リサイズされた画像のファイルパスとエラー情報
func (i Image) ResizeToFile() (string, error) {

	if i.Target == RATIO && i.ChangeRatio == 0 {
		return "", fmt.Errorf("invalid ratio")
	}
	if i.Target == WIDTH && i.ChangeWidth == 0 {
		return "", fmt.Errorf("invalid width")
	}
	if i.Target == HEIGHT && i.ChangeHeight == 0 {
		return "", fmt.Errorf("invalid height")
	}

	fileData, err := os.Open(i.FilePath)
	if err != nil {
		return "", err
	}

	// 画像をimage.Image型にdecodeします
	img, data, err := image.Decode(fileData)
	if err != nil {
		return "", err
	}
	err = fileData.Close()

	var width uint = 0
	var height uint = 0

	if i.Target == RATIO {
		width = uint(float32(img.Bounds().Dx()) * i.ChangeRatio)
	} else {
		if i.Target == WIDTH {
			width = uint(i.ChangeWidth)
		} else if i.Target == HEIGHT {
			height = uint(i.ChangeHeight)
		} else {
			return "", fmt.Errorf("invalid target")
		}
	}

	// 片方のサイズを0にするとアスペクト比固定してくれます
	resizedImg := resize.Resize(width, height, img, resize.NearestNeighbor)

	f, err := os.CreateTemp("", "gloudia.core.images.image.resize")
	if err != nil {
		return "", err
	}

	var destPath = f.Name()
	err = f.Close()
	if err != nil {
		return "", err
	}

	createFilePath := destPath + "." + data
	output, err := os.Create(createFilePath)
	if err != nil {
		return "", err
	}

	defer func() {
		err = output.Close()
		if err != nil {
			fmt.Printf("failed to close file: %v\n", err)
		}
	}()

	switch data {
	case "png":
		err = png.Encode(output, resizedImg)
	case "jpeg", "jpg":
		opts := &jpeg.Options{Quality: 100}
		err = jpeg.Encode(output, resizedImg, opts)
		/*
			case "gif":
				opts := &gif.Options{NumColors: 256}
				err = gif.Encode(output, resizedImg, opts)
		*/
	default:
		err = png.Encode(output, resizedImg)

	}
	if err != nil {
		return "", err
	}

	return createFilePath, nil
}