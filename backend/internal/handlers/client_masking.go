package handlers

import (
	"law-oa-go/internal/models"
	"law-oa-go/internal/services"
)

// maskClientResponseForC06 redacts the contact-bearing fields of a
// client-master response. Plan C06 option A grants business matter-management
// roles firm-wide basic reads, and those callers must receive the sensitive
// contact fields masked; matter-authorized lawyers keep their existing read
// semantics (contact person plaintext, pre-existing static phone masks).
// Edits go through the write endpoints, whose masked-echo guard keeps stored
// values intact when a masked response is echoed back.
func maskClientResponseForC06(response *services.ClientResponse) {
	if response == nil {
		return
	}
	response.ContactPerson = models.MaskContactPerson(response.ContactPerson)
	response.ContactPhone = models.MaskPhone(response.ContactPhone)
	response.Phone = models.MaskPhone(response.Phone)
}

// maskContactResponseForC06 applies the same business-management-role
// redaction to the independently stored primary-contact payload.
func maskContactResponseForC06(contact *services.ClientContactResponse) *services.ClientContactResponse {
	if contact == nil {
		return nil
	}
	masked := *contact
	masked.Name = models.MaskContactPerson(masked.Name)
	masked.Phone = models.MaskPhone(masked.Phone)
	masked.Email = models.MaskEmail(masked.Email)
	return &masked
}
