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
import { formatQuota, formatTimestampToDate } from '@/lib/format'
import { StatusBadge } from '@/components/status-badge'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { getRedemptionUses } from '../api'
import { type Redemption, type RedemptionUse } from '../types'

type RedemptionsUsesDialogProps = {
  open: boolean
  onOpenChange: (open: boolean) => void
  currentRow: Redemption
}

export function RedemptionsUsesDialog({
  open,
  onOpenChange,
  currentRow,
}: RedemptionsUsesDialogProps) {
  const { t } = useTranslation()
  const [uses, setUses] = useState<RedemptionUse[]>([])
  const [loading, setLoading] = useState(false)

  useEffect(() => {
    if (open && currentRow) {
      setLoading(true)
      getRedemptionUses(currentRow.id)
        .then((result) => {
          if (result.success && result.data) {
            setUses(result.data)
          }
        })
        .finally(() => setLoading(false))
    }
  }, [open, currentRow])

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className='sm:max-w-[600px]'>
        <DialogHeader>
          <DialogTitle>{t('Usage Records')}</DialogTitle>
          <DialogDescription>
            {t('Redemption code:')} {currentRow.name} ({currentRow.key.slice(0, 8)}...)
          </DialogDescription>
        </DialogHeader>

        <div className='space-y-4'>
          {/* 汇总信息 */}
          <div className='flex gap-4'>
            <StatusBadge
              label={`${t('Max Uses')}: ${currentRow.max_uses || 1}`}
              variant='neutral'
              copyable={false}
            />
            <StatusBadge
              label={`${t('Used')}: ${currentRow.used_count || 0}`}
              variant='neutral'
              copyable={false}
            />
            <StatusBadge
              label={`${t('Quota')}: ${formatQuota(currentRow.quota)}`}
              variant='neutral'
              copyable={false}
            />
          </div>

          {/* 使用记录列表 */}
          {loading ? (
            <div className='text-center py-8 text-muted-foreground'>
              {t('Loading...')}
            </div>
          ) : uses.length === 0 ? (
            <div className='text-center py-8 text-muted-foreground'>
              {t('No usage records')}
            </div>
          ) : (
            <div className='border rounded-lg overflow-hidden'>
              <table className='w-full text-sm'>
                <thead>
                  <tr className='border-b bg-muted/50'>
                    <th className='px-4 py-2 text-left font-medium'>
                      {t('User ID')}
                    </th>
                    <th className='px-4 py-2 text-left font-medium'>
                      {t('Quota')}
                    </th>
                    <th className='px-4 py-2 text-left font-medium'>
                      {t('Used Time')}
                    </th>
                  </tr>
                </thead>
                <tbody>
                  {uses.map((use) => (
                    <tr key={use.id} className='border-b last:border-0'>
                      <td className='px-4 py-2'>
                        <StatusBadge
                          label={t('User {{id}}', { id: use.user_id })}
                          variant='neutral'
                          copyable={false}
                        />
                      </td>
                      <td className='px-4 py-2 font-mono'>
                        {formatQuota(use.quota)}
                      </td>
                      <td className='px-4 py-2 font-mono'>
                        {formatTimestampToDate(use.used_time)}
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          )}
        </div>
      </DialogContent>
    </Dialog>
  )
}
