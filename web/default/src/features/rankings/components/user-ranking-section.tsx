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
import { useEffect, useMemo, useState } from 'react'
import { Users, Trophy } from 'lucide-react'
import { useTranslation } from 'react-i18next'
import { formatQuota } from '@/lib/format'
import { cn } from '@/lib/utils'
import { getUserRanking } from '@/features/user-ranking/api'
import type { UserRank } from '@/features/user-ranking/types'
import { formatTokens } from '../lib/format'
import type { RankingPeriod } from '../types'

/** Map rankings page period → user-ranking API period */
function toUserPeriod(period: RankingPeriod): string {
  switch (period) {
    case 'today':
      return 'day'
    case 'week':
      return 'week'
    case 'month':
      return 'month'
    case 'year':
      return 'all' // user-ranking API doesn't support 'year'
    case 'all':
      return 'all'
    default:
      return 'all'
  }
}

type UserRankingSectionProps = {
  period: RankingPeriod
}

export function UserRankingSection(props: UserRankingSectionProps) {
  const { t } = useTranslation()
  const [ranking, setRanking] = useState<UserRank[]>([])
  const [loading, setLoading] = useState(false)

  useEffect(() => {
    let cancelled = false
    const load = async () => {
      setLoading(true)
      try {
        const result = await getUserRanking({
          period: toUserPeriod(props.period) as 'day' | 'week' | 'month' | 'all',
          limit: 50,
        })
        if (!cancelled && result.success && result.data) {
          setRanking(result.data)
        }
      } finally {
        if (!cancelled) setLoading(false)
      }
    }
    load()
    return () => { cancelled = true }
  }, [props.period])

  const maxQuota = ranking.length > 0 ? ranking[0].quota_used : 1

  return (
    <section className='bg-card overflow-hidden rounded-lg border'>
      <header className='flex items-start justify-between gap-4 px-5 py-4'>
        <div className='min-w-0 flex-1'>
          <h2 className='text-foreground inline-flex items-center gap-2 text-base font-semibold'>
            <Users className='text-primary size-4' />
            {t('User Ranking')}
          </h2>
          <p className='text-muted-foreground mt-1 text-sm'>
            {t('Users ranked by usage volume')}
          </p>
        </div>
        <div className='shrink-0 text-right'>
          <div className='text-foreground font-mono text-2xl font-semibold tabular-nums'>
            {ranking.length}
          </div>
          <div className='text-muted-foreground/80 text-[10px] font-medium tracking-widest uppercase'>
            {t('Users')}
          </div>
        </div>
      </header>

      {/* Self rank card placeholder — only shown when logged in */}

      {/* Leaderboard */}
      <div className='border-t'>
        <header className='px-5 pt-4 pb-2'>
          <h3 className='text-foreground inline-flex items-center gap-2 text-sm font-semibold'>
            <Trophy className='size-3.5 text-amber-500' />
            {t('User Ranking')}
          </h3>
        </header>

        {loading ? (
          <div className='text-muted-foreground/80 px-5 py-8 text-center text-sm'>
            {t('Loading...')}
          </div>
        ) : ranking.length === 0 ? (
          <div className='text-muted-foreground/80 px-5 py-8 text-center text-sm'>
            {t('No data available')}
          </div>
        ) : (
          <div className='px-5 pb-4'>
            <div className='grid grid-cols-1 gap-x-8 md:grid-cols-2'>
              <UserList
                rows={ranking.slice(0, Math.ceil(ranking.length / 2))}
                maxQuota={maxQuota}
                rankOffset={0}
              />
              <UserList
                rows={ranking.slice(Math.ceil(ranking.length / 2))}
                maxQuota={maxQuota}
                rankOffset={Math.ceil(ranking.length / 2)}
              />
            </div>
          </div>
        )}
      </div>
    </section>
  )
}

function UserList(props: { rows: UserRank[]; maxQuota: number; rankOffset: number }) {
  const { t } = useTranslation()
  return (
    <ul>
      {props.rows.map((row, idx) => {
        const rank = props.rankOffset + idx + 1
        return (
          <li
            key={row.user_id}
            data-rank={rank}
            className='flex items-center gap-3 py-2.5 rounded-md transition-colors'
          >
            <span className='text-muted-foreground/80 w-6 shrink-0 text-right font-mono text-xs tabular-nums'>
              {rank}.
            </span>
            <div className='min-w-0 flex-1'>
              <span className='text-foreground block truncate text-sm font-medium'>
                {row.username}
              </span>
              <div className='ranking-progress-track'>
                <div
                  className='ranking-progress-bar'
                  style={{
                    width: `${Math.max(2, (row.quota_used / props.maxQuota) * 100)}%`,
                  }}
                />
              </div>
            </div>
            <div className='shrink-0 text-right'>
              <div className='text-foreground font-mono text-sm font-semibold tabular-nums'>
                {formatQuota(row.quota_used)}
              </div>
              <div className='text-muted-foreground/80 text-[11px]'>
                {row.request_count.toLocaleString()} {t('Requests')}
              </div>
            </div>
          </li>
        )
      })}
    </ul>
  )
}
