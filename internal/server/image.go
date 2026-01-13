package server

import (
	"camagru/internal/models"
	"database/sql"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	_ "image/png"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

func (s *Server) HandleCompose(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	user, err := s.GetCurrentUser(r)
	if err != nil {
		s.SendJSON(w, http.StatusUnauthorized, models.APIResponse{
			Success: false,
			Message: "Authentication required",
		})
		return
	}

	// Parse multipart form
	err = r.ParseMultipartForm(10 << 20) // 10MB
	if err != nil {
		s.SendJSON(w, http.StatusBadRequest, models.APIResponse{
			Success: false,
			Message: "Failed to parse form",
		})
		return
	}

	// Get uploaded image
	file, fileHeader, err := r.FormFile("image")
	if err != nil {
		s.SendJSON(w, http.StatusBadRequest, models.APIResponse{
			Success: false,
			Message: "No image file provided",
		})
		return
	}
	defer file.Close()

	// Validate filename (only allow safe characters: alphanumeric, underscore, hyphen, dot)
	filename := fileHeader.Filename
	if filename != "" {
		validFilename := true
		for _, char := range filename {
			if !((char >= 'a' && char <= 'z') ||
				(char >= 'A' && char <= 'Z') ||
				(char >= '0' && char <= '9') ||
				char == '_' || char == '-' || char == '.') {
				validFilename = false
				break
			}
		}
		if !validFilename {
			s.SendJSON(w, http.StatusBadRequest, models.APIResponse{
				Success: false,
				Message: "Please rename the file before uploading thanks!",
			})
			return
		}
	}

	// Validate MIME type (if provided)
	// Note: We also validate by attempting to decode the image, so MIME type is secondary
	contentType := fileHeader.Header.Get("Content-Type")
	if contentType != "" {
		validMIMETypes := []string{"image/jpeg", "image/jpg", "image/png", "image/gif", "image/webp"}
		validMIME := false
		for _, mime := range validMIMETypes {
			if contentType == mime {
				validMIME = true
				break
			}
		}
		if !validMIME {
			s.SendJSON(w, http.StatusBadRequest, models.APIResponse{
				Success: false,
				Message: "Invalid file type. Please upload an image file.",
			})
			return
		}
	}

	// Get asset ID
	assetIDStr := r.FormValue("asset_id")
	assetID, err := strconv.Atoi(assetIDStr)
	if err != nil || assetID == 0 {
		s.SendJSON(w, http.StatusBadRequest, models.APIResponse{
			Success: false,
			Message: "Invalid asset ID",
		})
		return
	}

	// Get overlay position and size (optional, defaults to center)
	overlayX := 0
	overlayY := 0
	overlayW := 0
	overlayH := 0
	if xStr := r.FormValue("overlay_x"); xStr != "" {
		overlayX, _ = strconv.Atoi(xStr)
	}
	if yStr := r.FormValue("overlay_y"); yStr != "" {
		overlayY, _ = strconv.Atoi(yStr)
	}
	if wStr := r.FormValue("overlay_w"); wStr != "" {
		overlayW, _ = strconv.Atoi(wStr)
	}
	if hStr := r.FormValue("overlay_h"); hStr != "" {
		overlayH, _ = strconv.Atoi(hStr)
	}

	// Get asset path
	var assetPath string
	err = s.DB.QueryRow("SELECT path FROM assets WHERE id = ?", assetID).Scan(&assetPath)
	if err == sql.ErrNoRows {
		s.SendJSON(w, http.StatusBadRequest, models.APIResponse{
			Success: false,
			Message: "Asset not found",
		})
		return
	}

	// Decode base image
	baseImg, _, err := image.Decode(file)
	if err != nil {
		s.SendJSON(w, http.StatusBadRequest, models.APIResponse{
			Success: false,
			Message: "Failed to decode image",
		})
		return
	}

	// Load overlay image
	// Asset path is like "/static/assets/cat.png", need to convert to "./web/static/assets/cat.png"
	overlayPath := "./web" + assetPath
	overlayFile, err := os.Open(overlayPath)
	if err != nil {
		// Try alternative path (without /web prefix)
		altPath := "." + assetPath
		overlayFile, err = os.Open(altPath)
		if err != nil {
			s.SendJSON(w, http.StatusInternalServerError, models.APIResponse{
				Success: false,
				Message: "Failed to load overlay image",
			})
			return
		}
	}
	defer overlayFile.Close()

	overlayImg, _, err := image.Decode(overlayFile)
	if err != nil {
		s.SendJSON(w, http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Message: "Failed to decode overlay image",
		})
		return
	}

	// Create RGBA canvas
	bounds := baseImg.Bounds()
	canvas := image.NewRGBA(bounds)
	draw.Draw(canvas, bounds, baseImg, image.Point{}, draw.Src)

	// Draw overlay with alpha blending
	overlayBounds := overlayImg.Bounds()
	
	// Use provided position/size if available, otherwise center and scale
	var finalX, finalY, finalW, finalH int
	if overlayW > 0 && overlayH > 0 {
		// Use provided position and size
		finalX = overlayX
		finalY = overlayY
		finalW = overlayW
		finalH = overlayH
	} else {
		// Default: center and make overlay half the size
		overlaySize := bounds.Dx()
		if overlaySize > bounds.Dy() {
			overlaySize = bounds.Dy()
		}
		overlaySize = overlaySize / 2
		finalX = (bounds.Dx() - overlaySize) / 2
		finalY = (bounds.Dy() - overlaySize) / 2
		finalW = overlaySize
		finalH = overlaySize
	}

	// Ensure overlay fits within bounds
	if finalX < 0 {
		finalX = 0
	}
	if finalY < 0 {
		finalY = 0
	}
	if finalX+finalW > bounds.Dx() {
		finalW = bounds.Dx() - finalX
	}
	if finalY+finalH > bounds.Dy() {
		finalH = bounds.Dy() - finalY
	}

	// Resize overlay to match desired size and draw
	if overlayBounds.Dx() != finalW || overlayBounds.Dy() != finalH {
		resizedOverlay := resizeImage(overlayImg, finalW, finalH)
		drawOverlay(canvas, resizedOverlay, finalX, finalY)
	} else {
		drawOverlay(canvas, overlayImg, finalX, finalY)
	}

	// Save composed image
	uploadDir := "./data/uploads"
	os.MkdirAll(uploadDir, 0755)

	// Generate a new filename for the saved image (overwrite the uploaded filename)
	filename = fmt.Sprintf("%d_%s_%d.jpg", user.ID, user.Username, time.Now().Unix())
	filePath := filepath.Join(uploadDir, filename)

	outFile, err := os.Create(filePath)
	if err != nil {
		s.SendJSON(w, http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Message: "Failed to save image",
		})
		return
	}
	defer outFile.Close()

	// Encode as JPEG
	err = jpeg.Encode(outFile, canvas, &jpeg.Options{Quality: 90})
	if err != nil {
		s.SendJSON(w, http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Message: "Failed to encode image",
		})
		return
	}

	// Save to database
	imagePath := "/static/uploads/" + filename
	_, err = s.DB.Exec(`
		INSERT INTO images (user_id, path)
		VALUES (?, ?)
	`, user.ID, imagePath)

	if err != nil {
		os.Remove(filePath)
		s.SendJSON(w, http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Message: "Failed to save image record",
		})
		return
	}

	s.SendJSON(w, http.StatusOK, models.APIResponse{
		Success: true,
		Message: "Image saved successfully",
		Data: map[string]interface{}{
			"path": imagePath,
		},
	})
}

func resizeImage(img image.Image, width, height int) image.Image {
	bounds := img.Bounds()
	srcW := bounds.Dx()
	srcH := bounds.Dy()

	dst := image.NewRGBA(image.Rect(0, 0, width, height))

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			srcX := x * srcW / width
			srcY := y * srcH / height
			dst.Set(x, y, img.At(bounds.Min.X+srcX, bounds.Min.Y+srcY))
		}
	}

	return dst
}

func drawOverlay(dst *image.RGBA, overlay image.Image, x, y int) {
	bounds := overlay.Bounds()
	dstBounds := dst.Bounds()

	for oy := 0; oy < bounds.Dy(); oy++ {
		for ox := 0; ox < bounds.Dx(); ox++ {
			dstX := x + ox
			dstY := y + oy

			if dstX < 0 || dstX >= dstBounds.Dx() || dstY < 0 || dstY >= dstBounds.Dy() {
				continue
			}

			srcColor := overlay.At(bounds.Min.X+ox, bounds.Min.Y+oy)
			dstColor := dst.At(dstX, dstY)

			// Alpha blending
			srcR, srcG, srcB, srcA := colorToRGBA(srcColor)
			dstR, dstG, dstB, dstA := colorToRGBA(dstColor)

			if srcA == 0 {
				continue // Fully transparent, skip
			}

			alpha := float64(srcA) / 255.0
			invAlpha := 1.0 - alpha

			r := uint8(float64(srcR)*alpha + float64(dstR)*invAlpha)
			g := uint8(float64(srcG)*alpha + float64(dstG)*invAlpha)
			b := uint8(float64(srcB)*alpha + float64(dstB)*invAlpha)
			a := uint8(255)

			if srcA < 255 {
				a = uint8(float64(srcA) + float64(dstA)*invAlpha)
			}

			dst.Set(dstX, dstY, color.RGBA{r, g, b, a})
		}
	}
}

func colorToRGBA(c color.Color) (r, g, b, a uint8) {
	r32, g32, b32, a32 := c.RGBA()
	return uint8(r32 >> 8), uint8(g32 >> 8), uint8(b32 >> 8), uint8(a32 >> 8)
}

