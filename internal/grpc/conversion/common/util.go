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

func ConvertList[in any, out any](src []*in, convertFunc func(*in) (*out, error)) ([]*out, error) {
	if len(src) == 0 {
		return nil, nil
	}
	dest := make([]*out, 0, len(src))
	for _, item := range src {
		if item == nil {
			continue
		}
		outItem, err := convertFunc(item)
		if err != nil {
			return nil, err
		}
		dest = append(dest, outItem)
	}
	return dest, nil
}

func ConvertToResponse[In any, Pb any, R any](
	in *In,
	convert func(*In) (*Pb, error),
	factory func(*Pb) *R,
) (*R, error) {
	pb, err := convert(in)
	if err != nil {
		return nil, err
	}
	return factory(pb), nil
}

func ConvertToListResponse[In any, Pb any, R any](
	in []*In,
	nextPageToken string,
	convert func([]*In) ([]*Pb, error),
	factory func([]*Pb, string) *R,
) (*R, error) {
	pb, err := convert(in)
	if err != nil {
		return nil, err
	}
	return factory(pb, nextPageToken), nil
}
