package models

import "time"

type GalleryConfigDTO struct {
	Title    string `json:"title"`
	Subtitle string `json:"subtitle"`
	Template string `json:"template"`
}

type RoomDTO struct {
	ID              uint   `json:"id"`
	Name            string `json:"name"`
	Description     string `json:"description"`
	Template        string `json:"template"`
	Width           int    `json:"width"`
	Depth           int    `json:"depth"`
	Height          int    `json:"height"`
	SortOrder       int    `json:"sort_order"`
	DoorNorthRoomID uint    `json:"door_north_room_id"`
	DoorEastRoomID  uint    `json:"door_east_room_id"`
	DoorSouthRoomID uint    `json:"door_south_room_id"`
	DoorWestRoomID  uint    `json:"door_west_room_id"`
	DoorNorthPosX   float64 `json:"door_north_pos_x"`
	DoorEastPosX    float64 `json:"door_east_pos_x"`
	DoorSouthPosX   float64 `json:"door_south_pos_x"`
	DoorWestPosX    float64 `json:"door_west_pos_x"`
}

type ArtworkDTO struct {
	ID                uint      `json:"id"`
	CreatedAt         time.Time `json:"created_at"`
	RoomID            uint      `json:"room_id"`
	SlideID           uint      `json:"slide_id"`
	ImageURL          string    `json:"image_url"`
	Wall              string    `json:"wall"`
	PosX              float64   `json:"pos_x"`
	PosY              float64   `json:"pos_y"`
	Scale             string    `json:"scale"`
	WidthPx           float64   `json:"width_px"`
	FrameStyle        string    `json:"frame_style"`
	Title             string    `json:"title"`
	Artist            string    `json:"artist"`
	Year              string    `json:"year"`
	Technique         string    `json:"technique"`
	Dimensions        string    `json:"dimensions"`
	Description       string    `json:"description"`
	ExternalLink      string    `json:"external_link"`
	ExternalLinkLabel string    `json:"external_link_label"`
}

type GalleryResponse struct {
	Config   GalleryConfigDTO `json:"config"`
	Rooms    []RoomDTO        `json:"rooms"`
	Artworks []ArtworkDTO     `json:"artworks"`
}
