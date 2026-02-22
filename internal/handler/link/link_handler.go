package link

import (
	"errors"
	"os"
	"strconv"

	"go-short/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type LinkHandler struct {
	linkService *service.LinkService
}

func NewLinkHandler(linkService *service.LinkService) *LinkHandler {
	return &LinkHandler{
		linkService: linkService,
	}
}

// Create 创建短链接
func (h *LinkHandler) Create(c *gin.Context) {
	var req CreateLinkRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, ErrInvalidRequest)
		return
	}

	uidStr := c.GetString("uid")
	userID, err := uuid.Parse(uidStr)
	if err != nil {
		// 处理无效 UUID
		c.JSON(400, ErrInvalidUserID)
		return
	}

	// 调用 Service 层处理业务逻辑
	cmd := service.CreateLinkCommand{
		OriginalURL: req.URL,
		Alias:       req.Alias,
		ShortCode:   req.ShortCode,
		Status:      req.Status,
		ExpiresAt:   req.ExpiresAt,
		UserID:      userID,
	}

	link, err := h.linkService.CreateLink(c, cmd)
	if err != nil {
		// 检查是否是已存在链接
		if errors.Is(err, service.ErrLinkAlreadyExists) {
			baseURL := os.Getenv("BASE_URL")
			if baseURL == "" {
				baseURL = "http://localhost:8080"
			}
			shortURL := baseURL + "/code/" + link.ShortCode
			c.JSON(200, NewLinkExistsResponse(link, shortURL))
			return
		}
		// 检查是否是短码重复
		if errors.Is(err, service.ErrShortCodeExists) {
			c.JSON(400, ErrShortCodeDuplicate)
			return
		}
		c.JSON(500, ErrDatabase)
		return
	}

	baseURL := os.Getenv("BASE_URL")
	if baseURL == "" {
		baseURL = "http://localhost:8080"
	}
	shortURL := baseURL + "/code/" + link.ShortCode

	c.JSON(200, NewCreateLinkResponse(link, shortURL))
}

// GetLinks 获取用户的链接列表
func (h *LinkHandler) GetLinks(c *gin.Context) {
	uidStr := c.GetString("uid")
	userID, err := uuid.Parse(uidStr)
	if err != nil {
		c.JSON(400, ErrInvalidUserID)
		return
	}
	pageStr := c.DefaultQuery("page", "1")
	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1 // 容错：非法值重置为第一页
	}

	list, total, err := h.linkService.GetLinksByUser(c, userID, page, 10)
	if err != nil {
		c.JSON(500, ErrDatabase)
		return
	}

	baseURL := os.Getenv("BASE_URL")
	if baseURL == "" {
		baseURL = "http://localhost:8080"
	}
	c.JSON(200, NewListLinksResponse(list, total, page, 10, baseURL))
}

// GetLinksByAlias 根据用户短链接别名获取链接列表
func (h *LinkHandler) GetLinksByAlias(c *gin.Context) {
	uidStr := c.GetString("uid")
	userID, err := uuid.Parse(uidStr)
	if err != nil {
		c.JSON(400, ErrInvalidUserID)
		return
	}
	pageStr := c.DefaultQuery("page", "1")
	alias := c.Query("alias")
	if alias == "" {
		c.JSON(400, ErrInvalidRequest)
		return
	}
	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1 // 容错：非法值重置为第一页
	}

	list, total, err := h.linkService.GetLinksByUserAlias(c, userID, alias, page, 10)
	if err != nil {
		c.JSON(500, ErrDatabase)
		return
	}

	baseURL := os.Getenv("BASE_URL")
	if baseURL == "" {
		baseURL = "http://localhost:8080"
	}
	c.JSON(200, NewListLinksResponse(list, total, page, 10, baseURL))
}

// Delete 根据ID删除短链接
func (h *LinkHandler) Delete(c *gin.Context) {
	linkIDStr := c.Param("id")
	linkID, err := strconv.ParseInt(linkIDStr, 10, 64)
	if err != nil || linkID <= 0 {
		c.JSON(400, ErrInvalidRequest)
		return
	}

	uidStr := c.GetString("uid")
	userID, err := uuid.Parse(uidStr)
	if err != nil {
		c.JSON(400, ErrInvalidUserID)
		return
	}
	isAdmin := c.GetString("role") == "admin"
	err = h.linkService.DeleteLink(c, linkID, userID, isAdmin)
	if err != nil {
		if errors.Is(err, service.ErrLinkNotFound) {
			c.JSON(404, ErrLinkNotFound)
			return
		}
		if errors.Is(err, service.ErrForbidden) {
			c.JSON(403, ErrForbidden)
			return
		}
		c.JSON(500, ErrDatabase)
		return
	}

	c.JSON(200, NewDeleteLinkResponse())
}
