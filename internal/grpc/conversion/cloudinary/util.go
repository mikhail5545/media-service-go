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

package cloudinary

import (
	"github.com/google/uuid"
	metaconv "github.com/mikhail5545/media-service-go/internal/grpc/conversion/cloudinary/metadata"
	"github.com/mikhail5545/media-service-go/internal/grpc/conversion/common"
	assetmodel "github.com/mikhail5545/media-service-go/internal/models/cloudinary/asset"
	bytesutil "github.com/mikhail5545/media-service-go/internal/util/bytes"
	cldassetpbv1 "github.com/mikhail5545/media-service-go/pb/media_service/cloudinary/asset/v1"
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
		logger:        logger.With(zap.String("component", "grpc/conversion/cloudinary")),
		metaConverter: metaconv.New(logger),
	}
}

func (c *Converter) convertUUIDs(asset *assetmodel.Asset, pbAsset *cldassetpbv1.Asset) (err error) {
	convert := func(id *uuid.UUID) ([]byte, error) {
		bytes, err := bytesutil.UUIDToBytes(id)
		if err != nil {
			c.logger.Error("failed to convert uuid", zap.Error(err))
			return nil, status.Error(codes.Internal, "failed to convert uuid")
		}
		return bytes, nil
	}

	pbAsset.Uuid, err = convert(&asset.ID)
	if err != nil {
		return err
	}
	pbAsset.CreatedBy, err = convert(asset.CreatedBy)
	if err != nil {
		return err
	}
	pbAsset.MarkedAsBrokenBy, err = convert(asset.MarkedAsBrokenBy)
	if err != nil {
		return err
	}
	pbAsset.RestoredBy, err = convert(asset.RestoredBy)
	if err != nil {
		return err
	}
	return nil
}

func (c *Converter) statusToProto(st assetmodel.Status) (cldassetpbv1.AssetStatus, error) {
	switch st {
	case assetmodel.StatusActive:
		return cldassetpbv1.AssetStatus_ASSET_STATUS_ACTIVE, nil
	case assetmodel.StatusUploadURLGenerated:
		return cldassetpbv1.AssetStatus_ASSET_STATUS_UPLOAD_URL_GENERATED, nil
	case assetmodel.StatusBroken:
		return cldassetpbv1.AssetStatus_ASSET_STATUS_BROKEN, nil
	case assetmodel.StatusArchived:
		return cldassetpbv1.AssetStatus_ASSET_STATUS_ARCHIVED, nil
	default:
		c.logger.Error("unknown asset status", zap.String("status", string(st)))
		return cldassetpbv1.AssetStatus_ASSET_STATUS_UNSPECIFIED, status.Error(codes.Internal, "unknown asset status")
	}
}

func (c *Converter) AssetToProto(asset *assetmodel.Asset) (*cldassetpbv1.Asset, error) {
	pbAsset := &cldassetpbv1.Asset{
		CreatedAt:            timestamppb.New(asset.CreatedAt),
		UpdatedAt:            timestamppb.New(asset.UpdatedAt),
		CloudinaryAssetId:    asset.CloudinaryAssetID,
		Url:                  asset.URL,
		SecureUrl:            asset.SecureURL,
		CloudinaryPublicId:   asset.CloudinaryPublicID,
		ResourceType:         asset.ResourceType,
		Format:               asset.Format,
		Tags:                 asset.Tags,
		CreatedByName:        asset.CreatedByName,
		MarkedAsBrokenByName: asset.MarkedAsBrokenByName,
		ArchivedByName:       asset.ArchivedByName,
		RestoredByName:       asset.RestoredByName,
		Note:                 asset.Note,
		ArchiveReason:        asset.ArchiveReason,
	}
	if asset.DeletedAt.Valid {
		pbAsset.DeletedAt = timestamppb.New(asset.DeletedAt.Time)
	}
	if err := c.convertUUIDs(asset, pbAsset); err != nil {
		return nil, err
	}
	pbStatus, err := c.statusToProto(asset.Status)
	if err != nil {
		return nil, err
	}
	pbAsset.Status = pbStatus
	return pbAsset, nil
}

func (c *Converter) DetailsToProto(details *assetmodel.Details) (*cldassetpbv1.Details, error) {
	asset, err := c.AssetToProto(details.Asset)
	if err != nil {
		return nil, err
	}
	meta, err := c.metaConverter.ToProto(details.Metadata)
	if err != nil {
		return nil, err
	}
	return &cldassetpbv1.Details{
		Asset:         asset,
		AssetMetadata: meta,
	}, nil
}

func (c *Converter) DetailsToProtoList(detailsList []*assetmodel.Details) ([]*cldassetpbv1.Details, error) {
	return common.ConvertList(detailsList, c.DetailsToProto)
}

type getRequest interface {
	GetUuid() []byte
}

func (c *Converter) ConvertGetRequest(req getRequest) (*assetmodel.GetFilter, error) {
	id, err := bytesutil.ToUUID(req.GetUuid())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid asset uuid")
	}
	return &assetmodel.GetFilter{
		ID: id.String(),
	}, nil
}

func (c *Converter) ConvertGetResponse(details *assetmodel.Details) (*cldassetpbv1.GetResponse, error) {
	return common.ConvertToResponse(details, c.DetailsToProto, func(pb *cldassetpbv1.Details) *cldassetpbv1.GetResponse {
		return &cldassetpbv1.GetResponse{
			Details: pb,
		}
	})
}

func (c *Converter) ConvertGetWithArchivedResponse(details *assetmodel.Details) (*cldassetpbv1.GetWithArchivedResponse, error) {
	return common.ConvertToResponse(details, c.DetailsToProto, func(pb *cldassetpbv1.Details) *cldassetpbv1.GetWithArchivedResponse {
		return &cldassetpbv1.GetWithArchivedResponse{
			Details: pb,
		}
	})
}

func (c *Converter) ConvertGetWithBrokenResponse(details *assetmodel.Details) (*cldassetpbv1.GetWithBrokenResponse, error) {
	return common.ConvertToResponse(details, c.DetailsToProto, func(pb *cldassetpbv1.Details) *cldassetpbv1.GetWithBrokenResponse {
		return &cldassetpbv1.GetWithBrokenResponse{
			Details: pb,
		}
	})
}

type listRequest interface {
	GetUuids() [][]byte
	GetCloudinaryAssetIds() []string
	GetCloudinaryPublicIds() []string
	GetResourceTypes() []string
	GetFormats() []string
	GetOrderDir() string
	GetOrderBy() string
	GetPageSize() int32
	GetPageToken() string
}

func (c *Converter) ConvertListRequest(req listRequest) (*assetmodel.ListRequest, error) {
	listReq := &assetmodel.ListRequest{
		CloudinaryAssetIDs:  req.GetCloudinaryAssetIds(),
		CloudinaryPublicIDs: req.GetCloudinaryPublicIds(),
		ResourceTypes:       req.GetResourceTypes(),
		Formats:             req.GetFormats(),
		OrderDir:            assetmodel.OrderDirection(req.GetOrderDir()),
		OrderField:          assetmodel.OrderField(req.GetOrderBy()),
		PageSize:            int(req.GetPageSize()),
		PageToken:           req.GetPageToken(),
	}
	ids, err := bytesutil.SliceToUUIDStrings(req.GetUuids())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid asset uuid in list")
	}
	listReq.IDs = ids
	return listReq, nil
}

func (c *Converter) ConvertListResponse(detailsList []*assetmodel.Details, nextPageToken string) (*cldassetpbv1.ListResponse, error) {
	return common.ConvertToListResponse(detailsList, nextPageToken, c.DetailsToProtoList,
		func(pbList []*cldassetpbv1.Details, token string) *cldassetpbv1.ListResponse {
			return &cldassetpbv1.ListResponse{
				Details:       pbList,
				NextPageToken: token,
			}
		})
}

func (c *Converter) ConvertListArchivedResponse(detailsList []*assetmodel.Details, nextPageToken string) (*cldassetpbv1.ListArchivedResponse, error) {
	return common.ConvertToListResponse(detailsList, nextPageToken, c.DetailsToProtoList,
		func(pbList []*cldassetpbv1.Details, token string) *cldassetpbv1.ListArchivedResponse {
			return &cldassetpbv1.ListArchivedResponse{
				Details:       pbList,
				NextPageToken: token,
			}
		})
}

func (c *Converter) ConvertListBrokenResponse(detailsList []*assetmodel.Details, nextPageToken string) (*cldassetpbv1.ListBrokenResponse, error) {
	return common.ConvertToListResponse(detailsList, nextPageToken, c.DetailsToProtoList,
		func(pbList []*cldassetpbv1.Details, token string) *cldassetpbv1.ListBrokenResponse {
			return &cldassetpbv1.ListBrokenResponse{
				Details:       pbList,
				NextPageToken: token,
			}
		})
}

type manageOwnerRequest interface {
	GetUuid() []byte
	GetOwnerUuid() []byte
	GetOwnerType() string
}

func (c *Converter) ConvertManageOwnerRequest(req manageOwnerRequest) (*assetmodel.ManageOwnerRequest, error) {
	assetID, err := bytesutil.ToUUID(req.GetUuid())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid asset uuid")
	}
	ownerID, err := bytesutil.ToUUID(req.GetOwnerUuid())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid owner uuid")
	}
	return &assetmodel.ManageOwnerRequest{
		ID:        assetID.String(),
		OwnerID:   ownerID.String(),
		OwnerType: req.GetOwnerType(),
	}, nil
}

type changeStateRequest interface {
	GetUuid() []byte
	GetAdminUuid() []byte
	GetNote() string
}

func (c *Converter) ConvertChangeStateRequest(req changeStateRequest) (*assetmodel.ChangeStateRequest, error) {
	assetID, err := bytesutil.ToUUID(req.GetUuid())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid asset uuid")
	}
	adminID, err := bytesutil.ToUUID(req.GetAdminUuid())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid admin uuid")
	}
	return &assetmodel.ChangeStateRequest{
		ID:      assetID.String(),
		AdminID: adminID.String(),
		Note:    req.GetNote(),
	}, nil
}

func (c *Converter) ConvertCreateSignedUploadURLRequest(req *cldassetpbv1.CreateSignedUploadURLRequest) (*assetmodel.CreateSignedUploadURLRequest, error) {
	uploadReq := &assetmodel.CreateSignedUploadURLRequest{
		Eager:     req.Eager,
		PublicID:  req.PublicId,
		File:      req.File,
		AdminName: req.AdminName,
		Note:      req.GetNote(),
	}
	adminID, err := bytesutil.ToUUID(req.GetAdminUuid())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid admin uuid")
	}
	uploadReq.AdminID = adminID.String()
	return uploadReq, nil
}

func (c *Converter) ConvertCreateSignedUploadURLResponse(uploadURL *assetmodel.GeneratedSignedParams) (*cldassetpbv1.CreateSignedUploadURLResponse, error) {
	return &cldassetpbv1.CreateSignedUploadURLResponse{
		Signature:    uploadURL.Signature,
		ApiKey:       uploadURL.ApiKey,
		Timestamp:    uploadURL.Timestamp,
		Eager:        uploadURL.Eager,
		PublicId:     uploadURL.PublicID,
		ResourceType: uploadURL.ResourceType,
	}, nil
}
