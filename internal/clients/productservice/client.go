// github.com/mikhail5545/product-service-go
// microservice for vitianmove project family
// Copyright (C) 2025  Mikhail Kulik

// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published
// by the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.

// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package productservice

import (
	"context"
	"fmt"
	"log"

	coursepb "github.com/mikhail5545/proto-go/proto/course/v0"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type CourseServiceClient struct {
	conn   *grpc.ClientConn
	client coursepb.CourseServiceClient
}

func NewCourseServiceClient(ctx context.Context, addr string) (*CourseServiceClient, error) {
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to gRPC server: %w", err)
	}
	log.Printf("gRPC connection to product (course) service at %s established", addr)

	client := coursepb.NewCourseServiceClient(conn)
	return &CourseServiceClient{
		conn:   conn,
		client: client,
	}, nil
}

func (c *CourseServiceClient) GetCoursePart(ctx context.Context, req *coursepb.GetCoursePartRequest) (*coursepb.GetCoursePartResponse, error) {
	return c.client.GetCoursePart(ctx, req)
}

func (c *CourseServiceClient) UpdateCoursePart(ctx context.Context, req *coursepb.UpdateCoursePartRequest) (*coursepb.UpdateCoursePartResponse, error) {
	return c.client.UpdateCoursePart(ctx, req)
}
