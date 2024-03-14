/*
 * Copyright (C) 2024 Nethesis S.r.l.
 * http://www.nethesis.it - info@nethesis.it
 *
 * SPDX-License-Identifier: GPL-2.0-only
 *
 * author: Edoardo Spadoni <edoardo.spadoni@nethesis.it>
 */

package models

import "time"

type AddRequest struct {
	UnitId string `json:"unit_id" binding:"required"`
}

type RegisterRequest struct {
	UnitId           string `json:"unit_id" binding:"required"`
	Username         string `json:"username" binding:"required"`
	Password         string `json:"password" binding:"required"`
	UnitName         string `json:"unit_name" binding:"required"`
	Version          string `json:"version"`
	SubscriptionType string `json:"subscription_type"`
	SystemId         string `json:"system_id"`
}

type Unit struct {
	ID               string    `json:"unit_id" structs:"id"`
	Name             string    `json:"unit_name" structs:"name"`
	Version          string    `json:"version" structs:"version"`
	SubscriptionType string    `json:"subscription_type" structs:"subscription_type"`
	SystemID         string    `json:"system_id" structs:"system_id"`
	Created          time.Time `json:"created" structs:"created"`
}
