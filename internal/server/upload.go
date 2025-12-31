package server

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/ydcloud-dy/opshub/internal/biz/rbac"
	rbaccustom "github.com/ydcloud-dy/opshub/internal/service/rbac"
	appLogger "github.com/ydcloud-dy/opshub/pkg/logger"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// UploadServer 上传服务
type UploadServer struct {
	db        *gorm.DB
	uploadDir string
	uploadURL string
}

// NewUploadServer 创建上传服务
func NewUploadServer(db *gorm.DB, uploadDir, uploadURL string) *UploadServer {
	// 确保上传目录存在
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		appLogger.Error("创建上传目录失败", zap.Error(err))
	}

	return &UploadServer{
		db:        db,
		uploadDir: uploadDir,
		uploadURL: uploadURL,
	}
}

// UploadAvatar 上传头像
func (s *UploadServer) UploadAvatar(c *gin.Context) {
	// 获取上传的文件
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(400, gin.H{
			"code":    400,
			"message": "获取文件失败",
		})
		return
	}
	defer file.Close()

	// 验证文件类型
	if !strings.HasPrefix(header.Header.Get("Content-Type"), "image/") {
		c.JSON(400, gin.H{
			"code":    400,
			"message": "只能上传图片文件",
		})
		return
	}

	// 验证文件大小 (2MB)
	if header.Size > 2*1024*1024 {
		c.JSON(400, gin.H{
			"code":    400,
			"message": "图片大小不能超过 2MB",
		})
		return
	}

	// 获取当前用户ID (从JWT中获取)
	userID := rbaccustom.GetUserID(c)
	if userID == 0 {
		c.JSON(401, gin.H{
			"code":    401,
			"message": "未登录",
		})
		return
	}

	// 生成唯一文件名
	ext := filepath.Ext(header.Filename)
	filename := fmt.Sprintf("avatar_%d_%d%s", userID, time.Now().Unix(), ext)

	// 保存文件
	dst := filepath.Join(s.uploadDir, filename)
	out, err := os.Create(dst)
	if err != nil {
		appLogger.Error("创建文件失败", zap.Error(err))
		c.JSON(500, gin.H{
			"code":    500,
			"message": "保存文件失败",
		})
		return
	}
	defer out.Close()

	_, err = io.Copy(out, file)
	if err != nil {
		appLogger.Error("写入文件失败", zap.Error(err))
		c.JSON(500, gin.H{
			"code":    500,
			"message": "保存文件失败",
		})
		return
	}

	// 生成访问URL
	fileURL := fmt.Sprintf("%s/%s", s.uploadURL, filename)

	appLogger.Info("头像上传成功",
		zap.Uint("userID", userID),
		zap.String("filename", filename),
		zap.String("url", fileURL),
	)

	c.JSON(200, gin.H{
		"code":    0,
		"message": "success",
		"data": gin.H{
			"url":  fileURL,
			"path": filename,
		},
	})
}

// UpdateUserAvatar 更新用户头像
func (s *UploadServer) UpdateUserAvatar(c *gin.Context) {
	var req struct {
		Avatar string `json:"avatar" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{
			"code":    400,
			"message": "参数错误",
		})
		return
	}

	// 获取当前用户ID
	userID := rbaccustom.GetUserID(c)
	if userID == 0 {
		c.JSON(401, gin.H{
			"code":    401,
			"message": "未登录",
		})
		return
	}

	// 更新用户头像
	if err := s.db.Model(&rbac.SysUser{}).Where("id = ?", userID).Update("avatar", req.Avatar).Error; err != nil {
		appLogger.Error("更新头像失败",
			zap.Uint("userID", userID),
			zap.Error(err),
		)
		c.JSON(500, gin.H{
			"code":    500,
			"message": "更新头像失败",
		})
		return
	}

	appLogger.Info("用户头像更新成功",
		zap.Uint("userID", userID),
		zap.String("avatar", req.Avatar),
	)

	c.JSON(200, gin.H{
		"code":    0,
		"message": "头像更新成功",
	})
}
