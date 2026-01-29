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

package common

import (
	"context"

	errutil "github.com/mikhail5545/media-service-go/internal/util/errors"
)

func HandleList[Req any, InternalReq any, InternalRes any, Res any](
	ctx context.Context,
	toInternal func(Req) (InternalReq, error),
	toRes func([]InternalRes, string) (*Res, error),
	fn func(context.Context, InternalReq) ([]InternalRes, string, error),
	req Req,
) (*Res, error) {
	converted, err := toInternal(req)
	if err != nil {
		return nil, err
	}
	internalRes, nextPageToken, err := fn(ctx, converted)
	if err != nil {
		return nil, errutil.ToGRPCCode(err)
	}
	return toRes(internalRes, nextPageToken)
}

func HandleEmpty[Req any, Internal any, Res any](
	ctx context.Context,
	convFunc func(Req) (Internal, error),
	fn func(context.Context, Internal) error,
	req Req,
	res *Res,
) (*Res, error) {
	converted, err := convFunc(req)
	if err != nil {
		return nil, err
	}
	if err := fn(ctx, converted); err != nil {
		return nil, errutil.ToGRPCCode(err)
	}
	return res, nil
}

func Handle[Req any, Internal any, InternalRes any, Res any](
	ctx context.Context,
	toInternal func(Req) (Internal, error),
	toProto func(InternalRes) (*Res, error),
	fn func(context.Context, Internal) (InternalRes, error),
	req Req,
) (*Res, error) {
	internal, err := toInternal(req)
	if err != nil {
		return nil, err
	}
	internalRes, err := fn(ctx, internal)
	if err != nil {
		return nil, errutil.ToGRPCCode(err)
	}
	return toProto(internalRes)
}
