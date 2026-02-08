package audit

import (
	sharedAPI "easi/backend/internal/shared/api"
)

type AuditLinks struct {
	*sharedAPI.HATEOASLinks
}

func NewAuditLinks(h *sharedAPI.HATEOASLinks) *AuditLinks {
	return &AuditLinks{HATEOASLinks: h}
}

func (h *AuditLinks) AuditHistory(id string) string {
	return h.Base() + "/audit/" + id
}
