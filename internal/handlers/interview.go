package handlers

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sort"

	"biviz/internal/auth"
	"biviz/internal/middleware"
)

const (
	uploadDir      = "uploads/interviews"
	maxUploadBytes = 50 << 20 // 50MB
)

type uploadedFile struct {
	Name string
	Size string
	Time string
}

func ShowInterviewUpload(w http.ResponseWriter, r *http.Request) {
	session := middleware.GetCurrentSession(r)
	if session == nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	user, err := auth.GetUserByID(r.Context(), session.UserID)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	files := listUploadedFiles(session.UserID)

	render(w, "interview-upload.html", map[string]interface{}{
		"User":  user,
		"Files": files,
	})
}

func HandleInterviewUpload(w http.ResponseWriter, r *http.Request) {
	session := middleware.GetCurrentSession(r)
	if session == nil {
		http.Error(w, "인증이 필요합니다", http.StatusUnauthorized)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, maxUploadBytes)
	if err := r.ParseMultipartForm(maxUploadBytes); err != nil {
		http.Error(w, "파일이 너무 큽니다 (최대 50MB)", http.StatusBadRequest)
		return
	}

	userDir := filepath.Join(uploadDir, fmt.Sprintf("%d", session.UserID))
	if err := os.MkdirAll(userDir, 0755); err != nil {
		http.Error(w, "서버 오류", http.StatusInternalServerError)
		return
	}

	fhs := r.MultipartForm.File["files"]
	if len(fhs) == 0 {
		http.Error(w, "파일을 선택해주세요", http.StatusBadRequest)
		return
	}

	for _, fh := range fhs {
		src, err := fh.Open()
		if err != nil {
			continue
		}

		dst, err := os.Create(filepath.Join(userDir, filepath.Base(fh.Filename)))
		if err != nil {
			src.Close()
			continue
		}

		io.Copy(dst, src)
		src.Close()
		dst.Close()
	}

	w.WriteHeader(http.StatusOK)
}

func listUploadedFiles(userID int) []uploadedFile {
	userDir := filepath.Join(uploadDir, fmt.Sprintf("%d", userID))
	entries, err := os.ReadDir(userDir)
	if err != nil {
		return nil
	}

	var files []uploadedFile
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		info, err := e.Info()
		if err != nil {
			continue
		}
		files = append(files, uploadedFile{
			Name: e.Name(),
			Size: formatBytes(info.Size()),
			Time: info.ModTime().Format("2006-01-02 15:04"),
		})
	}

	sort.Slice(files, func(i, j int) bool {
		return files[i].Time > files[j].Time
	})

	return files
}

func formatBytes(b int64) string {
	switch {
	case b < 1024:
		return fmt.Sprintf("%d B", b)
	case b < 1024*1024:
		return fmt.Sprintf("%.1f KB", float64(b)/1024)
	default:
		return fmt.Sprintf("%.1f MB", float64(b)/(1024*1024))
	}
}
