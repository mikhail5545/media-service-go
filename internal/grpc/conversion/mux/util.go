/*
 * Copyright (c) 2026. Mikhail Kulik
 *
 * This program is free software: you can redistribute it and/or modify
 *  it under the terms of the GNU Affero General Public License as published
 *  by the Free Software Foundation, either version 3 of the License, or
 *  (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 *  but WITHOUT ANY WARRANTY; without even the implied warranty of
 *  MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 *  GNU Affero General Public License for more details.
 *
 * You should have received a copy of the GNU Affero General Public License
 *  along with this program.  If not, see <https://www.gnu.org/licenses/>.
 */

package mux

import (
	"github.com/google/uuid"
	"github.com/mikhail5545/media-service-go/internal/grpc/conversion/common"
	metaconv "github.com/mikhail5545/media-service-go/internal/grpc/conversion/mux/metadata"
	assetmodel "github.com/mikhail5545/media-service-go/internal/models/mux/asset"
	bytesutil "github.com/mikhail5545/media-service-go/internal/util/bytes"
	muxassetpbv1 "github.com/mikhail5545/media-service-go/pb/media_service/mux/asset/v1"
	muxgo "github.com/muxinc/mux-go/v6"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Converter struct {
	logger        *zap.Logger
	metaConverter *metaconv.Converter
}

func New(logger *zap.Logger) *Converter {
	return &Converter{
		logger:        logger.With(zap.String("component", "grpc/mux/Converter")),
		metaConverter: metaconv.New(logger),
	}
}

func convertUUIDs(asset *assetmodel.Asset, pbAsset *muxassetpbv1.Asset) error {
	convert := func(id *uuid.UUID) ([]byte, error) {
		bytes, err := bytesutil.UUIDToBytes(id)
		if err != nil {
			return nil, status.Error(codes.Internal, "failed to convert uuid to bytes")
		}
		return bytes, nil
	}

	var err error
	pbAsset.Uuid, err = convert(&asset.ID)
	if err != nil {
		return err
	}
	pbAsset.CreatedBy, err = convert(asset.CreatedBy)
	if err != nil {
		return err
	}
	pbAsset.ArchivedBy, err = convert(asset.ArchivedBy)
	if err != nil {
		return err
	}
	pbAsset.RestoredBy, err = convert(asset.RestoredBy)
	if err != nil {
		return err
	}
	pbAsset.MarkedAsBrokenBy, err = convert(asset.MarkedAsBrokenBy)
	if err != nil {
		return err
	}
	return nil
}

func statusToProto(st assetmodel.Status, logger *zap.Logger) (muxassetpbv1.AssetStatus, error) {
	switch st {
	case assetmodel.StatusUploadURLGenerated:
		return muxassetpbv1.AssetStatus_ASSET_STATUS_UPLOAD_URL_GENERATED, nil
	case assetmodel.StatusActive:
		return muxassetpbv1.AssetStatus_ASSET_STATUS_ACTIVE, nil
	case assetmodel.StatusArchived:
		return muxassetpbv1.AssetStatus_ASSET_STATUS_ARCHIVED, nil
	case assetmodel.StatusBroken:
		return muxassetpbv1.AssetStatus_ASSET_STATUS_BROKEN, nil
	default:
		logger.Error("unknown asset status", zap.String("status", string(st)))
		return muxassetpbv1.AssetStatus_ASSET_STATUS_UNSPECIFIED, status.Error(codes.Internal, "unknown asset status")
	}
}

func ingestTypeToProto(it assetmodel.IngestType, logger *zap.Logger) (muxassetpbv1.AssetIngestType, error) {
	switch it {
	case assetmodel.IngestTypeLiveSRT:
		return muxassetpbv1.AssetIngestType_ASSET_INGEST_TYPE_LIVE_SRT, nil
	case assetmodel.IngestTypeLiveRTMP:
		return muxassetpbv1.AssetIngestType_ASSET_INGEST_TYPE_LIVE_RTMP, nil
	case assetmodel.IngestTypeOnDemandClip:
		return muxassetpbv1.AssetIngestType_ASSET_INGEST_TYPE_ON_DEMAND_CLIP, nil
	case assetmodel.IngestTypeOnDemandDirectUpload:
		return muxassetpbv1.AssetIngestType_ASSET_INGEST_TYPE_ON_DEMAND_DIRECT_UPLOAD, nil
	case assetmodel.IngestTypeOnDemandURL:
		return muxassetpbv1.AssetIngestType_ASSET_INGEST_TYPE_ON_DEMAND_URL, nil
	default:
		logger.Error("unknown asset ingest type", zap.String("ingest_type", string(it)))
		return muxassetpbv1.AssetIngestType_ASSET_INGEST_TYPE_UNSPECIFIED, status.Error(codes.Internal, "unknown asset ingest type")
	}
}

func protoToIngestType(it muxassetpbv1.AssetIngestType) (assetmodel.IngestType, error) {
	switch it {
	case muxassetpbv1.AssetIngestType_ASSET_INGEST_TYPE_LIVE_SRT:
		return assetmodel.IngestTypeLiveSRT, nil
	case muxassetpbv1.AssetIngestType_ASSET_INGEST_TYPE_LIVE_RTMP:
		return assetmodel.IngestTypeLiveRTMP, nil
	case muxassetpbv1.AssetIngestType_ASSET_INGEST_TYPE_ON_DEMAND_CLIP:
		return assetmodel.IngestTypeOnDemandClip, nil
	case muxassetpbv1.AssetIngestType_ASSET_INGEST_TYPE_ON_DEMAND_DIRECT_UPLOAD:
		return assetmodel.IngestTypeOnDemandDirectUpload, nil
	case muxassetpbv1.AssetIngestType_ASSET_INGEST_TYPE_ON_DEMAND_URL:
		return assetmodel.IngestTypeOnDemandURL, nil
	default:
		return "", status.Error(codes.InvalidArgument, "unknown asset ingest type")
	}
}

func protoToIngestTypes(its []muxassetpbv1.AssetIngestType) ([]assetmodel.IngestType, error) {
	ingestTypes := make([]assetmodel.IngestType, 0, len(its))
	for _, pbIt := range its {
		it, err := protoToIngestType(pbIt)
		if err != nil {
			return nil, err
		}
		ingestTypes = append(ingestTypes, it)
	}
	return ingestTypes, nil
}

func stateToProto(state assetmodel.State, logger *zap.Logger) (muxassetpbv1.AssetState, error) {
	switch state {
	case assetmodel.StateTranscoding:
		return muxassetpbv1.AssetState_ASSET_STATE_TRANSCODING, nil
	case assetmodel.StateIngesting:
		return muxassetpbv1.AssetState_ASSET_STATE_INGESTING, nil
	case assetmodel.StateCompleted:
		return muxassetpbv1.AssetState_ASSET_STATE_COMPLETED, nil
	case assetmodel.StateLive:
		return muxassetpbv1.AssetState_ASSET_STATE_LIVE, nil
	case assetmodel.StateErrored:
		return muxassetpbv1.AssetState_ASSET_STATE_ERRORED, nil
	default:
		logger.Error("unknown asset state", zap.String("state", string(state)))
		return muxassetpbv1.AssetState_ASSET_STATE_UNSPECIFIED, status.Error(codes.Internal, "unknown asset state")
	}
}

func uploadStatusToProto(us assetmodel.UploadStatus, logger *zap.Logger) (muxassetpbv1.AssetUploadStatus, error) {
	switch us {
	case assetmodel.UploadStatusPreparing:
		return muxassetpbv1.AssetUploadStatus_ASSET_UPLOAD_STATUS_PREPARING, nil
	case assetmodel.UploadStatusReady:
		return muxassetpbv1.AssetUploadStatus_ASSET_UPLOAD_STATUS_READY, nil
	case assetmodel.UploadStatusErrored:
		return muxassetpbv1.AssetUploadStatus_ASSET_UPLOAD_STATUS_ERRORED, nil
	case assetmodel.UploadStatusDeleted:
		return muxassetpbv1.AssetUploadStatus_ASSET_UPLOAD_STATUS_DELETED, nil
	default:
		logger.Error("unknown asset upload status", zap.String("upload_status", string(us)))
		return muxassetpbv1.AssetUploadStatus_ASSET_UPLOAD_STATUS_UNSPECIFIED, status.Error(codes.Internal, "unknown asset upload status")
	}
}

func protoToUploadStatus(us muxassetpbv1.AssetUploadStatus) (assetmodel.UploadStatus, error) {
	switch us {
	case muxassetpbv1.AssetUploadStatus_ASSET_UPLOAD_STATUS_PREPARING:
		return assetmodel.UploadStatusPreparing, nil
	case muxassetpbv1.AssetUploadStatus_ASSET_UPLOAD_STATUS_READY:
		return assetmodel.UploadStatusReady, nil
	case muxassetpbv1.AssetUploadStatus_ASSET_UPLOAD_STATUS_ERRORED:
		return assetmodel.UploadStatusErrored, nil
	case muxassetpbv1.AssetUploadStatus_ASSET_UPLOAD_STATUS_DELETED:
		return assetmodel.UploadStatusDeleted, nil
	default:
		return "", status.Error(codes.Internal, "unknown asset upload status")
	}
}

func protoToUploadStatuses(us []muxassetpbv1.AssetUploadStatus) ([]assetmodel.UploadStatus, error) {
	uploadStatuses := make([]assetmodel.UploadStatus, 0, len(us))
	for _, pbUs := range us {
		uStatus, err := protoToUploadStatus(pbUs)
		if err != nil {
			return nil, err
		}
		uploadStatuses = append(uploadStatuses, uStatus)
	}
	return uploadStatuses, nil
}

func enumValuesToProto(asset *assetmodel.Asset, pbAsset *muxassetpbv1.Asset, logger *zap.Logger) error {
	var err error
	pbAsset.Status, err = statusToProto(asset.Status, logger)
	if err != nil {
		return err
	}

	pbAsset.IngestType, err = ingestTypeToProto(asset.IngestType, logger)
	if err != nil {
		return err
	}

	pbAsset.State, err = stateToProto(asset.State, logger)
	if err != nil {
		return err
	}

	pbAsset.UploadStatus, err = uploadStatusToProto(asset.UploadStatus, logger)
	if err != nil {
		return err
	}

	return nil
}

func (c *Converter) AssetToProto(asset *assetmodel.Asset) (*muxassetpbv1.Asset, error) {
	pbAsset := &muxassetpbv1.Asset{
		CreatedAt:               timestamppb.New(asset.CreatedAt),
		UpdatedAt:               timestamppb.New(asset.UpdatedAt),
		MuxUploadId:             asset.MuxUploadID,
		MuxAssetId:              asset.MuxAssetID,
		AspectRatio:             asset.AspectRatio,
		Duration:                asset.Duration,
		CreatedByName:           asset.CreatedByName,
		ArchivedByName:          asset.ArchivedByName,
		RestoredByName:          asset.RestoredByName,
		MarkedAsBrokenByName:    asset.MarkedAsBrokenByName,
		Note:                    asset.Note,
		ArchiveReason:           asset.ArchiveReason,
		ArchiveEventId:          asset.ArchiveEventID,
		PrimarySignedPlaybackId: asset.PrimarySignedPlaybackID,
		PrimaryPublicPlaybackId: asset.PrimaryPublicPlaybackID,
	}
	if asset.DeletedAt.Valid {
		pbAsset.DeletedAt = timestamppb.New(asset.DeletedAt.Time)
	}
	if err := convertUUIDs(asset, pbAsset); err != nil {
		return nil, err
	}
	if err := enumValuesToProto(asset, pbAsset, c.logger); err != nil {
		return nil, err
	}
	return pbAsset, nil
}

func (c *Converter) DetailsToProto(details *assetmodel.Details) (*muxassetpbv1.Details, error) {
	asset, err := c.AssetToProto(details.Asset)
	if err != nil {
		return nil, err
	}
	meta, err := c.metaConverter.ToProto(details.Metadata)
	if err != nil {
		return nil, err
	}
	return &muxassetpbv1.Details{
		Asset:         asset,
		AssetMetadata: meta,
	}, nil
}

func (c *Converter) DetailsToProtoList(details []*assetmodel.Details) ([]*muxassetpbv1.Details, error) {
	return common.ConvertList(details, func(d *assetmodel.Details) (*muxassetpbv1.Details, error) {
		return c.DetailsToProto(d)
	})
}

type getRequest interface {
	GetUuid() []byte
	GetUploadStatus() muxassetpbv1.AssetUploadStatus
}

func (c *Converter) ConvertGetRequest(req getRequest) (*assetmodel.GetFilter, error) {
	filter := &assetmodel.GetFilter{}
	id, err := bytesutil.ToUUID(req.GetUuid())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid asset uuid: %v", err)
	}
	filter.ID = id.String()
	uploadStatus, err := protoToUploadStatus(req.GetUploadStatus())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid upload status: %v", err)
	}
	filter.UploadStatus = uploadStatus
	return filter, nil
}

func (c *Converter) ConvertGetResponse(details *assetmodel.Details) (*muxassetpbv1.GetResponse, error) {
	return common.ConvertToResponse(details, c.DetailsToProto,
		func(pb *muxassetpbv1.Details) *muxassetpbv1.GetResponse {
			return &muxassetpbv1.GetResponse{
				Details: pb,
			}
		})
}

func (c *Converter) ConvertGetWithArchivedResponse(details *assetmodel.Details) (*muxassetpbv1.GetWithArchivedResponse, error) {
	return common.ConvertToResponse(details, c.DetailsToProto,
		func(pb *muxassetpbv1.Details) *muxassetpbv1.GetWithArchivedResponse {
			return &muxassetpbv1.GetWithArchivedResponse{
				Details: pb,
			}
		})
}

func (c *Converter) ConvertGetWithBrokenResponse(details *assetmodel.Details) (*muxassetpbv1.GetWithBrokenResponse, error) {
	return common.ConvertToResponse(details, c.DetailsToProto,
		func(pb *muxassetpbv1.Details) *muxassetpbv1.GetWithBrokenResponse {
			return &muxassetpbv1.GetWithBrokenResponse{
				Details: pb,
			}
		})
}

type listRequest interface {
	GetUuids() [][]byte
	GetMuxUploadIds() []string
	GetMuxAssetIds() []string
	GetAspectRatios() []string
	GetResolutionTiers() []string
	GetIngestTypes() []muxassetpbv1.AssetIngestType
	GetUploadStatuses() []muxassetpbv1.AssetUploadStatus
	GetOrderBy() string
	GetOrderDir() string
	GetPageSize() int32
	GetNextPageToken() string
}

func (c *Converter) ConvertListRequest(req listRequest) (*assetmodel.ListRequest, error) {
	listReq := &assetmodel.ListRequest{
		MuxUploadIDs:    req.GetMuxUploadIds(),
		MuxAssetIDs:     req.GetMuxAssetIds(),
		AspectRatios:    req.GetAspectRatios(),
		ResolutionTiers: req.GetResolutionTiers(),
		PageSize:        int(req.GetPageSize()),
		PageToken:       req.GetNextPageToken(),
		OrderBy:         assetmodel.OrderField(req.GetOrderBy()),
		OrderDir:        assetmodel.OrderDirection(req.GetOrderDir()),
	}
	ids, err := bytesutil.SliceToUUIDStrings(req.GetUuids())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid asset uuids: %v", err)
	}
	listReq.IDs = ids

	pbIngestTypes := req.GetIngestTypes()
	if len(pbIngestTypes) > 0 {
		ingestTypes, err := protoToIngestTypes(pbIngestTypes)
		if err != nil {
			return nil, err
		}
		listReq.IngestTypes = ingestTypes
	}

	pbUploadStatuses := req.GetUploadStatuses()
	if len(pbUploadStatuses) > 0 {
		uploadStatuses, err := protoToUploadStatuses(pbUploadStatuses)
		if err != nil {
			return nil, err
		}
		listReq.UploadStatuses = uploadStatuses
	}

	return listReq, nil
}

func (c *Converter) ConvertListResponse(details []*assetmodel.Details, nextPageToken string) (*muxassetpbv1.ListResponse, error) {
	return common.ConvertToListResponse(details, nextPageToken, c.DetailsToProtoList,
		func(pbList []*muxassetpbv1.Details, token string) *muxassetpbv1.ListResponse {
			return &muxassetpbv1.ListResponse{
				Details:       pbList,
				NextPageToken: token,
			}
		})
}

func (c *Converter) ConvertListArchivedResponse(details []*assetmodel.Details, nextPageToken string) (*muxassetpbv1.ListArchivedResponse, error) {
	return common.ConvertToListResponse(details, nextPageToken, c.DetailsToProtoList,
		func(pbList []*muxassetpbv1.Details, token string) *muxassetpbv1.ListArchivedResponse {
			return &muxassetpbv1.ListArchivedResponse{
				Details:       pbList,
				NextPageToken: token,
			}
		})
}

func (c *Converter) ConvertListBrokenResponse(details []*assetmodel.Details, nextPageToken string) (*muxassetpbv1.ListBrokenResponse, error) {
	return common.ConvertToListResponse(details, nextPageToken, c.DetailsToProtoList,
		func(pbList []*muxassetpbv1.Details, token string) *muxassetpbv1.ListBrokenResponse {
			return &muxassetpbv1.ListBrokenResponse{
				Details:       pbList,
				NextPageToken: token,
			}
		})
}

type changeStateRequest interface {
	GetUuid() []byte
	GetAdminUuid() []byte
	GetAdminName() string
	GetNote() string
}

func (c *Converter) ConvertChangeStateRequest(req changeStateRequest) (*assetmodel.ChangeStateRequest, error) {
	changeStateReq := &assetmodel.ChangeStateRequest{
		AdminName: req.GetAdminName(),
		Note:      req.GetNote(),
	}
	id, err := bytesutil.ToUUID(req.GetUuid())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid asset uuid: %v", err)
	}
	changeStateReq.ID = id.String()
	adminID, err := bytesutil.ToUUID(req.GetAdminUuid())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid admin uuid: %v", err)
	}
	changeStateReq.AdminID = adminID.String()
	return changeStateReq, nil
}

type manageOwnerRequest interface {
	GetUuid() []byte
	GetOwnerUuid() []byte
	GetOwnerType() string
}

func (c *Converter) ConvertManageOwnerRequest(req manageOwnerRequest) (*assetmodel.ManageOwnerRequest, error) {
	manageOwnerReq := &assetmodel.ManageOwnerRequest{
		OwnerType: req.GetOwnerType(),
	}
	id, err := bytesutil.ToUUID(req.GetUuid())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid asset uuid: %v", err)
	}
	manageOwnerReq.ID = id.String()
	ownerID, err := bytesutil.ToUUID(req.GetOwnerUuid())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid owner uuid: %v", err)
	}
	manageOwnerReq.OwnerID = ownerID.String()
	return manageOwnerReq, nil
}

func (c *Converter) ConvertCreateUploadURLRequest(req *muxassetpbv1.CreateUploadURLRequest) (*assetmodel.CreateUploadURLRequest, error) {
	createReq := &assetmodel.CreateUploadURLRequest{
		Title:     req.GetTitle(),
		AdminName: req.GetAdminName(),
	}
	adminID, err := bytesutil.ToUUID(req.GetAdminUuid())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid admin uuid: %v", err)
	}
	createReq.AdminID = adminID.String()
	return createReq, nil
}

func (c *Converter) ConvertCreateUploadURLResponse(data *muxgo.UploadResponse) (*muxassetpbv1.CreateUploadURLResponse, error) {
	return &muxassetpbv1.CreateUploadURLResponse{
		Url:        data.Data.Url,
		Timeout:    int64(data.Data.Timeout),
		Status:     data.Data.Status,
		Id:         data.Data.Id,
		CorsOrigin: data.Data.CorsOrigin,
	}, nil
}

func (c *Converter) ConvertGeneratePlaybackTokenRequest(req *muxassetpbv1.GeneratePlaybackTokenRequest) (*assetmodel.GeneratePlaybackTokenRequest, error) {
	assetID, err := bytesutil.ToUUID(req.GetAssetUuid())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid asset uuid: %v", err)
	}
	userID, err := bytesutil.ToUUID(req.GetUserUuid())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid user uuid: %v", err)
	}
	sessionID, err := bytesutil.ToUUID(req.GetSessionUuid())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid session uuid: %v", err)
	}
	return &assetmodel.GeneratePlaybackTokenRequest{
		AssetID:    assetID,
		UserID:     userID,
		SessionID:  &sessionID,
		Expiration: req.GetExpiration(),
		UserAgent:  req.UserAgent,
	}, nil
}

func (c *Converter) ConvertGeneratePlaybackTokenResponse(token string) *muxassetpbv1.GeneratePlaybackTokenResponse {
	return &muxassetpbv1.GeneratePlaybackTokenResponse{
		Token: token,
	}
}
