package grpc

import (
	"context"
	"fmt"
	"time"

	"github.com/sachinthra/file-locker/backend/internal/storage"
	pb "github.com/sachinthra/file-locker/backend/pkg/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type FileServiceServer struct {
	pb.UnimplementedFileServiceServer
	redisCache *storage.RedisCache
}

func NewFileServiceServer(redisCache *storage.RedisCache) *FileServiceServer {
	return &FileServiceServer{
		redisCache: redisCache,
	}
}

func (s *FileServiceServer) GetFileMetadata(ctx context.Context, req *pb.FileRequest) (*pb.FileMetadata, error) {
	// Validate request
	if req.FileId == "" {
		return nil, status.Error(codes.InvalidArgument, "file_id is required")
	}
	if req.UserId == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}

	// Get metadata from Redis
	metadata, err := s.redisCache.GetFileMetadata(ctx, req.FileId)
	if err != nil {
		return nil, status.Error(codes.NotFound, "file not found")
	}

	// Verify ownership
	if metadata.UserID != req.UserId {
		return nil, status.Error(codes.PermissionDenied, "access denied")
	}

	// Convert to protobuf message
	pbMetadata := &pb.FileMetadata{
		FileId:        metadata.FileID,
		UserId:        metadata.UserID,
		FileName:      metadata.FileName,
		MimeType:      metadata.MimeType,
		Size:          metadata.Size,
		EncryptedSize: metadata.EncryptedSize,
		CreatedAt:     metadata.CreatedAt.Format(time.RFC3339),
		Tags:          metadata.Tags,
		DownloadCount: int32(metadata.DownloadCount),
	}

	if metadata.ExpiresAt != nil {
		pbMetadata.ExpiresAt = metadata.ExpiresAt.Format(time.RFC3339)
	}

	return pbMetadata, nil
}

func (s *FileServiceServer) ListFiles(ctx context.Context, req *pb.ListRequest) (*pb.FileList, error) {
	// Validate request
	if req.UserId == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}

	// Get user's file IDs
	fileIDs, err := s.redisCache.GetUserFiles(ctx, req.UserId)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to retrieve files")
	}

	// Collect metadata for all files
	files := make([]*pb.FileMetadata, 0)
	now := time.Now()

	for _, fileID := range fileIDs {
		metadata, err := s.redisCache.GetFileMetadata(ctx, fileID)
		if err != nil {
			continue // Skip files that no longer exist
		}

		// Filter out expired files
		if metadata.ExpiresAt != nil && metadata.ExpiresAt.Before(now) {
			continue
		}

		pbMetadata := &pb.FileMetadata{
			FileId:        metadata.FileID,
			UserId:        metadata.UserID,
			FileName:      metadata.FileName,
			MimeType:      metadata.MimeType,
			Size:          metadata.Size,
			EncryptedSize: metadata.EncryptedSize,
			CreatedAt:     metadata.CreatedAt.Format(time.RFC3339),
			Tags:          metadata.Tags,
			DownloadCount: int32(metadata.DownloadCount),
		}

		if metadata.ExpiresAt != nil {
			pbMetadata.ExpiresAt = metadata.ExpiresAt.Format(time.RFC3339)
		}

		files = append(files, pbMetadata)
	}

	// Apply pagination if requested
	page := int(req.Page)
	limit := int(req.Limit)

	if limit == 0 {
		limit = 100 // Default limit
	}
	if page == 0 {
		page = 1 // Default to first page
	}

	start := (page - 1) * limit
	end := start + limit

	total := len(files)

	if start >= total {
		return &pb.FileList{
			Files: []*pb.FileMetadata{},
			Total: int32(total),
		}, nil
	}

	if end > total {
		end = total
	}

	pagedFiles := files[start:end]

	return &pb.FileList{
		Files: pagedFiles,
		Total: int32(total),
	}, nil
}

func (s *FileServiceServer) UpdateTags(ctx context.Context, req *pb.UpdateTagsRequest) (*pb.FileMetadata, error) {
	// Validate request
	if req.FileId == "" {
		return nil, status.Error(codes.InvalidArgument, "file_id is required")
	}
	if req.UserId == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}

	// Get existing metadata
	metadata, err := s.redisCache.GetFileMetadata(ctx, req.FileId)
	if err != nil {
		return nil, status.Error(codes.NotFound, "file not found")
	}

	// Verify ownership
	if metadata.UserID != req.UserId {
		return nil, status.Error(codes.PermissionDenied, "access denied")
	}

	// Update tags
	metadata.Tags = req.Tags

	// Save updated metadata
	var expiration time.Duration
	if metadata.ExpiresAt != nil {
		expiration = time.Until(*metadata.ExpiresAt)
	}

	if err := s.redisCache.SaveFileMetadata(ctx, req.FileId, metadata, expiration); err != nil {
		return nil, status.Error(codes.Internal, "failed to update tags")
	}

	// Return updated metadata
	pbMetadata := &pb.FileMetadata{
		FileId:        metadata.FileID,
		UserId:        metadata.UserID,
		FileName:      metadata.FileName,
		MimeType:      metadata.MimeType,
		Size:          metadata.Size,
		EncryptedSize: metadata.EncryptedSize,
		CreatedAt:     metadata.CreatedAt.Format(time.RFC3339),
		Tags:          metadata.Tags,
		DownloadCount: int32(metadata.DownloadCount),
	}

	if metadata.ExpiresAt != nil {
		pbMetadata.ExpiresAt = metadata.ExpiresAt.Format(time.RFC3339)
	}

	return pbMetadata, nil
}

func (s *FileServiceServer) SetExpiration(ctx context.Context, req *pb.ExpirationRequest) (*pb.FileMetadata, error) {
	// Validate request
	if req.FileId == "" {
		return nil, status.Error(codes.InvalidArgument, "file_id is required")
	}
	if req.UserId == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}

	// Get existing metadata
	metadata, err := s.redisCache.GetFileMetadata(ctx, req.FileId)
	if err != nil {
		return nil, status.Error(codes.NotFound, "file not found")
	}

	// Verify ownership
	if metadata.UserID != req.UserId {
		return nil, status.Error(codes.PermissionDenied, "access denied")
	}

	// Parse expiration time
	if req.ExpiresAt != "" {
		expiresAt, err := time.Parse(time.RFC3339, req.ExpiresAt)
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("invalid expires_at format: %v", err))
		}
		metadata.ExpiresAt = &expiresAt
	} else {
		metadata.ExpiresAt = nil // Remove expiration
	}

	// Calculate Redis expiration
	var redisExpiration time.Duration
	if metadata.ExpiresAt != nil {
		redisExpiration = time.Until(*metadata.ExpiresAt) + 24*time.Hour
	}

	// Save updated metadata
	if err := s.redisCache.SaveFileMetadata(ctx, req.FileId, metadata, redisExpiration); err != nil {
		return nil, status.Error(codes.Internal, "failed to update expiration")
	}

	// Return updated metadata
	pbMetadata := &pb.FileMetadata{
		FileId:        metadata.FileID,
		UserId:        metadata.UserID,
		FileName:      metadata.FileName,
		MimeType:      metadata.MimeType,
		Size:          metadata.Size,
		EncryptedSize: metadata.EncryptedSize,
		CreatedAt:     metadata.CreatedAt.Format(time.RFC3339),
		Tags:          metadata.Tags,
		DownloadCount: int32(metadata.DownloadCount),
	}

	if metadata.ExpiresAt != nil {
		pbMetadata.ExpiresAt = metadata.ExpiresAt.Format(time.RFC3339)
	}

	return pbMetadata, nil
}
