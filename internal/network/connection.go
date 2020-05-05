/*
 * Copyright (c) 2020 ErikPelli <https://github.com/ErikPelli>
 * This file is part of GoombaGram.
 *
 * GoombaGram is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Affero General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 * GoombaGram is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Affero General Public License for more details.
 * You should have received a copy of the GNU Affero General Public License
 * along with GoombaGram.  If not, see <http://www.gnu.org/licenses/>.
 */

package network

type netInterface interface {
	Connect(address string, obfuscation bool) error
	Send(data []byte) error
	Receive(data []byte) error
	Close() error
}

var modes = []string {"abridged", "full", "intermediate", "intermediatePadded"}

// TCP (obfuscated mode available)
// 0: Abridged
// 1: Intermediate
// 2: PaddedIntermediate
//
// Full transport is useless because its features are already implemented in TCP protocol