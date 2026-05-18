/*
Copyright (C) 2023-2026 QuantumNous

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as
published by the Free Software Foundation, either version 3 of the
License, or (at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program. If not, see <https://www.gnu.org/licenses/>.

For commercial licensing, please contact support@quantumnous.com
*/
import { z } from 'zod'

// ============================================================================
// User Ranking Types
// ============================================================================

export const userRankSchema = z.object({
  user_id: z.number(),
  username: z.string(),
  quota_used: z.number(),
  request_count: z.number(),
  prompt_tokens: z.number(),
  completion_tokens: z.number(),
  total_tokens: z.number(),
})

export type UserRank = z.infer<typeof userRankSchema>

export type RankingPeriod = 'day' | 'week' | 'month' | 'all'

export interface GetRankingParams {
  period?: RankingPeriod
  limit?: number
}

export interface GetRankingResponse {
  success: boolean
  message?: string
  data?: UserRank[]
}

export interface GetSelfRankResponse {
  success: boolean
  message?: string
  data?: {
    rank: UserRank
    position: number
  }
}
