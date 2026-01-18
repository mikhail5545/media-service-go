/*
 * Copyright (c) 2026. Mikhail Kulik.
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Affero General Public License as published
 * by the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Affero General Public License for more details.
 *
 * You should have received a copy of the GNU Affero General Public License
 * along with this program.  If not, see <https://www.gnu.org/licenses/>.
 */

package patch

// UpdateIfChanged checks if new and cur are different. If they are, it adds the new value to the updates map under the dbField key.
func UpdateIfChanged[T comparable](updates map[string]any, dbField string, new, cur *T) {
	if new == nil {
		return
	}
	if cur == nil {
		updates[dbField] = *new
		return
	}
	if *cur != *new {
		updates[dbField] = *new
	}
}
