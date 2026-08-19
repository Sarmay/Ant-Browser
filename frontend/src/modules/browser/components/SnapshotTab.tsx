import { useEffect, useState } from 'react'
import { Archive, RotateCcw, Trash2 } from 'lucide-react'
import { Button, Card, ConfirmModal, Input, Table, toast } from '../../../shared/components'
import type { TableColumn } from '../../../shared/components/Table'
import type { SnapshotInfo } from '../types'
import { createSnapshot, deleteSnapshot, listSnapshots, restoreSnapshot } from '../api'

interface Props {
  profileId: string
  running: boolean
}

const defaultName = () => {
  const now = new Date()
  const pad = (n: number) => String(n).padStart(2, '0')
  return `快照_${now.getFullYear()}-${pad(now.getMonth() + 1)}-${pad(now.getDate())}_${pad(now.getHours())}:${pad(now.getMinutes())}`
}

const formatSize = (mb: number) => (mb < 1 ? `${(mb * 1024).toFixed(0)} KB` : `${mb.toFixed(1)} MB`)

const formatTime = (value: string) => {
  const d = new Date(value)
  return Number.isNaN(d.getTime()) ? '-' : d.toLocaleString('zh-CN')
}

export function SnapshotTab({ profileId, running }: Props) {
  const [snapshots, setSnapshots] = useState<SnapshotInfo[]>([])
  const [loading, setLoading] = useState(false)
  const [creating, setCreating] = useState(false)
  const [newName, setNewName] = useState(defaultName)
  const [confirmRestore, setConfirmRestore] = useState<string | null>(null)
  const [confirmDelete, setConfirmDelete] = useState<string | null>(null)
  const [actionLoading, setActionLoading] = useState<string | null>(null)
  const snapshotActionRunning = actionLoading !== null

  const load = async () => {
    setLoading(true)
    try {
      setSnapshots(await listSnapshots(profileId))
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => { load() }, [profileId])

  const handleCreate = async () => {
    if (!newName.trim() || running || creating || snapshotActionRunning) return
    setCreating(true)
    try {
      await createSnapshot(profileId, newName.trim())
      toast.success('快照创建成功')
      setNewName(defaultName())
      await load()
    } catch {
      toast.error('快照创建失败')
    } finally {
      setCreating(false)
    }
  }

  const handleRestore = async (snapshotId: string): Promise<boolean> => {
    if (running || creating || snapshotActionRunning) return false
    setActionLoading(snapshotId)
    try {
      await restoreSnapshot(profileId, snapshotId)
      toast.success('快照恢复成功')
      return true
    } catch (error: any) {
      toast.error(error?.message || '快照恢复失败')
      return false
    } finally {
      setActionLoading(null)
    }
  }

  const handleDelete = async (snapshotId: string): Promise<boolean> => {
    if (creating || snapshotActionRunning) return false
    setActionLoading(snapshotId)
    try {
      await deleteSnapshot(profileId, snapshotId)
      toast.success('快照已删除')
      try {
        await load()
      } catch (error: any) {
        toast.warning(error?.message || '快照已删除，但刷新列表失败，请手动刷新')
      }
      return true
    } catch (error: any) {
      toast.error(error?.message || '快照删除失败')
      return false
    } finally {
      setActionLoading(null)
    }
  }

  const columns: TableColumn<SnapshotInfo>[] = [
    { key: 'name', title: '名称' },
    { key: 'sizeMB', title: '大小', render: v => formatSize(v as number) },
    { key: 'createdAt', title: '创建时间', render: v => formatTime(v as string) },
    {
      key: 'snapshotId',
      title: '操作',
      render: (snapshotId) => {
        const sid = snapshotId as string
        return (
          <div className="flex gap-2">
            <Button
              size="sm"
              variant="ghost"
              disabled={running || snapshotActionRunning}
              title={running ? '请先停止实例' : snapshotActionRunning ? '正在处理快照操作' : '恢复此快照'}
              onClick={() => setConfirmRestore(sid)}
            >
              <RotateCcw className="w-3.5 h-3.5" />
              恢复
            </Button>
            <Button
              size="sm"
              variant="ghost"
              disabled={snapshotActionRunning}
              onClick={() => setConfirmDelete(sid)}
            >
              <Trash2 className="w-3.5 h-3.5" />
              删除
            </Button>
          </div>
        )
      },
    },
  ]

  return (
    <div className="space-y-4">
      <Card title="创建快照" subtitle={running ? '实例运行中，请先停止后再创建快照' : '将当前用户数据目录压缩为快照'}>
        <div className="flex flex-col sm:flex-row gap-3">
          <Input
            value={newName}
            onChange={e => setNewName(e.target.value)}
            placeholder="快照名称"
            disabled={running || creating || snapshotActionRunning}
            className="flex-1"
          />
          <Button onClick={handleCreate} disabled={running || creating || snapshotActionRunning || !newName.trim()}>
            <Archive className="w-4 h-4" />
            {creating ? '创建中...' : '创建快照'}
          </Button>
        </div>
      </Card>

      <Card
        title="快照列表"
        subtitle={loading ? '加载中...' : `共 ${snapshots.length} 个快照`}
      >
        {snapshots.length === 0 && !loading ? (
          <p className="text-sm text-[var(--color-text-muted)] py-6 text-center">暂无快照</p>
        ) : (
          <Table columns={columns} data={snapshots} rowKey="snapshotId" />
        )}
      </Card>
      <ConfirmModal
        open={confirmRestore !== null}
        onClose={() => setConfirmRestore(null)}
        onConfirm={() => confirmRestore ? handleRestore(confirmRestore) : false}
        title="恢复快照"
        content="恢复快照会覆盖当前实例的用户数据。此操作不可撤销，是否继续？"
        confirmText="恢复"
        danger
      />
      <ConfirmModal
        open={confirmDelete !== null}
        onClose={() => setConfirmDelete(null)}
        onConfirm={() => confirmDelete ? handleDelete(confirmDelete) : false}
        title="删除快照"
        content="确定删除这个快照？此操作不可恢复。"
        confirmText="删除"
        danger
      />
    </div>
  )
}
