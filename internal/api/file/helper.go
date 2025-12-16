package file

import (
	"net/http"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"os"
	"os/exec"
	
	"github.com/buckket/go-blurhash"
)

func detectMimeType(data []byte) string {
	return http.DetectContentType(data)
}

func classifyMediaType(mime string) string {
	if len(mime) == 0 {
		return "document"
	}
	switch {
	case mime[:6] == "image/":
		return "image"
	case mime[:6] == "video/":
		return "video"
	case mime[:6] == "audio/":
		return "audio"
	default:
		return "document"
	}
}

func generateBlurHash(img image.Image) (string, int, int, error) {
	width := img.Bounds().Dx()
	height := img.Bounds().Dy()

	hash, err := blurhash.Encode(4, 3, img)
	return hash, width, height, err
}

func extractVideoThumbnail(videoPath string) (string, error) {
	tmpFile, err := os.CreateTemp("", "thumb-*.jpg")
	if err != nil {
		return "", err
	}
	tmpFile.Close()

	cmd := exec.Command(
		"ffmpeg",
		"-y",
		"-i", videoPath,
		"-ss", "0.5",
		"-vframes", "1",
		"-vf", "scale=320:-1",
		tmpFile.Name(),
	)

	err = cmd.Run()
	if err != nil {
		return "", err
	}

	return tmpFile.Name(), nil
}

func blurHashFromVideo(videoPath string) (string, int, int, error) {
	thumbPath, err := extractVideoThumbnail(videoPath)
	if err != nil {
		return "", 0, 0, err
	}
	defer os.Remove(thumbPath)

	f, err := os.Open(thumbPath)
	if err != nil {
		return "", 0, 0, err
	}
	defer f.Close()

	img, _, err := image.Decode(f)
	if err != nil {
		return "", 0, 0, err
	}

	hash, err := blurhash.Encode(4, 3, img)
	if err != nil {
		return "", 0, 0, err
	}

	w := img.Bounds().Dx()
	h := img.Bounds().Dy()

	return hash, w, h, nil
}
