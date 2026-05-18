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
import { useEffect, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { formatQuota } from '@/lib/format'
import { StatusBadge } from '@/components/status-badge'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { getUserRanking, getSelfRank } from '../api'
import { type UserRank, type RankingPeriod } from '../types'

// ============================================================================
// User Ranking Component
// ============================================================================

export function UserRanking() {
  const { t } = useTranslation()
  const [period, setPeriod] = useState<RankingPeriod>('all')
  const [ranking, setRanking] = useState<UserRank[]>([])
  const [selfRank, setSelfRank] = useState<{
    rank: UserRank
    position: number
  } | null>(null)
  const [loading, setLoading] = useState(false)

  useEffect(() => {
    loadRanking()
    loadSelfRank()
  }, [period])

  const loadRanking = async () => {
    setLoading(true)
    try {
      const result = await getUserRanking({ period, limit: 50 })
      if (result.success && result.data) {
        setRanking(result.data)
      }
    } finally {
      setLoading(false)
    }
  }

  const loadSelfRank = async () => {
    const result = await getSelfRank(period)
    if (result.success && result.data) {
      setSelfRank(result.data)
    }
  }

  const getRankBadge = (position: number) => {
    if (position === 1) {
      return (
        <StatusBadge
          label='🥇'
          variant='success'
          copyable={false}
          className='text-lg'
        />
      )
    }
    if (position === 2) {
      return (
        <StatusBadge
          label='🥈'
          variant='neutral'
          copyable={false}
          className='text-lg'
        />
      )
    }
    if (position === 3) {
      return (
        <StatusBadge
          label='🥉'
          variant='warning'
          copyable={false}
          className='text-lg'
        />
      )
    }
    return (
      <StatusBadge
        label={`#${position}`}
        variant='neutral'
        copyable={false}
      />
    )
  }

  const formatTokens = (tokens: number) => {
    if (tokens >= 1000000) {
      return `${(tokens / 1000000).toFixed(1)}M`
    }
    if (tokens >= 1000) {
      return `${(tokens / 1000).toFixed(1)}K`
    }
    return tokens.toString()
  }

  return (
    <div className='space-y-4'>
      {/* 当前用户排名 */}
      {selfRank && (
        <Card>
          <CardHeader>
            <CardTitle>{t('Your Ranking')}</CardTitle>
          </CardHeader>
          <CardContent>
            <div className='flex items-center gap-4'>
              <div className='text-4xl font-bold'>#{selfRank.position}</div>
              <div className='space-y-1'>
                <div className='text-sm text-muted-foreground'>
                  {t('Total Usage')}
                </div>
                <div className='text-2xl font-semibold'>
                  {formatQuota(selfRank.rank.quota_used)}
                </div>
              </div>
              <div className='ml-auto space-y-1 text-right'>
                <div className='text-sm text-muted-foreground'>
                  {t('Requests')}
                </div>
                <div className='text-lg font-medium'>
                  {selfRank.rank.request_count.toLocaleString()}
                </div>
              </div>
            </div>
          </CardContent>
        </Card>
      )}

      {/* 排行榜 */}
      <Card>
        <CardHeader>
          <CardTitle>{t('User Ranking')}</CardTitle>
        </CardHeader>
        <CardContent>
          <Tabs value={period} onValueChange={(v) => setPeriod(v as RankingPeriod)}>
            <TabsList>
              <TabsTrigger value='day'>{t('Today')}</TabsTrigger>
              <TabsTrigger value='week'>{t('This Week')}</TabsTrigger>
              <TabsTrigger value='month'>{t('This Month')}</TabsTrigger>
              <TabsTrigger value='all'>{t('All Time')}</TabsTrigger>
            </TabsList>

            <TabsContent value={period} className='mt-4'>
              {loading ? (
                <div className='text-center py-8 text-muted-foreground'>
                  {t('Loading...')}
                </div>
              ) : ranking.length === 0 ? (
                <div className='text-center py-8 text-muted-foreground'>
                  {t('No data available')}
                </div>
              ) : (
                <div className='border rounded-lg overflow-hidden'>
                  <table className='w-full text-sm'>
                    <thead>
                      <tr className='border-b bg-muted/50'>
                        <th className='px-4 py-3 text-left font-medium w-[80px]'>
                          {t('Rank')}
                        </th>
                        <th className='px-4 py-3 text-left font-medium'>
                          {t('User')}
                        </th>
                        <th className='px-4 py-3 text-right font-medium'>
                          {t('Usage')}
                        </th>
                        <th className='px-4 py-3 text-right font-medium hidden md:table-cell'>
                          {t('Requests')}
                        </th>
                        <th className='px-4 py-3 text-right font-medium hidden lg:table-cell'>
                          {t('Tokens')}
                        </th>
                      </tr>
                    </thead>
                    <tbody>
                      {ranking.map((rank, index) => (
                        <tr
                          key={rank.user_id}
                          className='border-b last:border-0 hover:bg-muted/30'
                        >
                          <td className='px-4 py-3'>
                            {getRankBadge(index + 1)}
                          </td>
                          <td className='px-4 py-3 font-medium'>
                            {rank.username}
                          </td>
                          <td className='px-4 py-3 text-right font-mono'>
                            {formatQuota(rank.quota_used)}
                          </td>
                          <td className='px-4 py-3 text-right font-mono hidden md:table-cell'>
                            {rank.request_count.toLocaleString()}
                          </td>
                          <td className='px-4 py-3 text-right font-mono hidden lg:table-cell'>
                            {formatTokens(rank.total_tokens)}
                          </td>
                        </tr>
                      ))}
                    </tbody>
                  </table>
                </div>
              )}
            </TabsContent>
          </Tabs>
        </CardContent>
      </Card>
    </div>
  )
}
