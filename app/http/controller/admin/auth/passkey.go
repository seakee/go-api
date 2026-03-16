package auth

import (
	"errors"
	"io"

	"github.com/gin-gonic/gin"
	"github.com/seakee/go-api/app/pkg/e"
	"github.com/seakee/go-api/app/service/system"
)

func (h handler) BeginPasskeyRegistration() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			DisplayName string `json:"display_name"`
		}

		errCode := e.InvalidParams
		var err error
		var result system.PasskeyOptionsResult

		userID, _ := c.Get("user_id")
		if err = c.ShouldBindJSON(&req); errors.Is(err, io.EOF) || err == nil {
			result, errCode, err = h.service.BeginPasskeyRegistration(h.Context(c), userID.(uint), req.DisplayName)
		}

		h.I18n.JSON(c, errCode, result, err)
	}
}

func (h handler) FinishPasskeyRegistration() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			ChallengeID string                   `json:"challenge_id" binding:"required"`
			Credential  system.PasskeyCredential `json:"credential" binding:"required"`
		}

		errCode := e.InvalidParams
		var err error
		var result system.PasskeyItem

		userID, _ := c.Get("user_id")
		if err = c.ShouldBindJSON(&req); err == nil {
			result, errCode, err = h.service.FinishPasskeyRegistration(h.Context(c), userID.(uint), req.ChallengeID, req.Credential)
		}

		h.I18n.JSON(c, errCode, result, err)
	}
}

func (h handler) BeginPasskeyLogin() gin.HandlerFunc {
	return func(c *gin.Context) {
		errCode := e.InvalidParams
		var err error
		var result system.PasskeyOptionsResult

		if err = c.ShouldBindJSON(&struct{}{}); errors.Is(err, io.EOF) || err == nil {
			result, errCode, err = h.service.BeginPasskeyLogin(h.Context(c))
		}

		h.I18n.JSON(c, errCode, result, err)
	}
}

func (h handler) FinishPasskeyLogin() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			ChallengeID string                   `json:"challenge_id" binding:"required"`
			Credential  system.PasskeyCredential `json:"credential" binding:"required"`
		}

		errCode := e.InvalidParams
		var err error
		var result system.AccessToken

		if err = c.ShouldBindJSON(&req); err == nil {
			result, errCode, err = h.service.FinishPasskeyLogin(h.Context(c), req.ChallengeID, req.Credential)
		}

		h.I18n.JSON(c, errCode, result, err)
	}
}

func (h handler) Passkeys() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, _ := c.Get("user_id")
		list, errCode, err := h.service.ListPasskeys(h.Context(c), userID.(uint))
		h.I18n.JSON(c, errCode, gin.H{"list": list}, err)
	}
}

func (h handler) DeletePasskey() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			ID           uint   `json:"id" binding:"required"`
			ReauthTicket string `json:"reauth_ticket" binding:"required"`
		}

		errCode := e.InvalidParams
		var err error

		userID, _ := c.Get("user_id")
		if err = c.ShouldBindJSON(&req); err == nil {
			errCode, err = h.service.DeletePasskey(h.Context(c), userID.(uint), req.ID, req.ReauthTicket)
		}

		h.I18n.JSON(c, errCode, nil, err)
	}
}
