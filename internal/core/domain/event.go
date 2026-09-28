package core_domain

// Events are published by services and delivered to clients by the WebSocket hub.

type SpotUpdatedEvent struct {
	SpotId       string
	PresentCount int
}

type AnnouncementCreatedEvent struct {
	Announcement Announcement
}
