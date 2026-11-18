package file

import "github.com/google/uuid"

type FileMetadata struct {
	ID   uuid.UUID
	Name string
}
