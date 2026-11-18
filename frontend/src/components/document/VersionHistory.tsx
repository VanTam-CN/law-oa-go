import React, { useEffect, useState } from 'react'
import {
  Button,
  List,
  ListItem,
  ListItemText,
  ListItemSecondaryAction,
  Dialog,
  DialogTitle,
  DialogContent,
  DialogActions,
  Typography,
  Chip,
  Box,
  IconButton,
  Tooltip,
  CircularProgress,
  Alert,
  Divider,
} from '@mui/material'
import {
  History as HistoryIcon,
  Restore as RestoreIcon,
  Delete as DeleteIcon,
  Compare as CompareIcon,
  Download as DownloadIcon,
  Visibility as VisibilityIcon,
} from '@mui/icons-material'
import { formatDistanceToNow } from 'date-fns'
import { zhCN } from 'date-fns/locale'
import documentVersionService, {
  DocumentVersion,
  VersionComparison,
  LockStatus,
} from '@/services/documentVersion'

// ============================================================================
// 类型定义
// ============================================================================

interface VersionHistoryProps {
  documentId: number
  documentName?: string
  onClose?: () => void
  onVersionRestored?: (version: number) => void
  readOnly?: boolean
}

interface CompareDialogProps {
  open: boolean
  comparison: VersionComparison | null
  onClose: () => void
}

// ============================================================================
// 组件
// ============================================================================

/**
 * 版本历史组件
 */
export const VersionHistory: React.FC<VersionHistoryProps> = ({
  documentId,
  documentName,
  onClose,
  onVersionRestored,
  readOnly = false,
}) => {
  const [versions, setVersions] = useState<DocumentVersion[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [page, setPage] = useState(1)
  const [totalPages, setTotalPages] = useState(1)
  const [lockStatus, setLockStatus] = useState<LockStatus | null>(null)

  // 对话框状态
  const [restoreDialog, setRestoreDialog] = useState<{ open: boolean; version: number }>({
    open: false,
    version: 0,
  })
  const [deleteDialog, setDeleteDialog] = useState<{ open: boolean; version: number }>({
    open: false,
    version: 0,
  })
  const [compareDialog, setCompareDialog] = useState<CompareDialogProps>({
    open: false,
    comparison: null,
    onClose: () => setCompareDialog({ open: false, comparison: null }),
  })

  // 加载版本列表
  const loadVersions = async (currentPage = 1) => {
    setLoading(true)
    setError(null)

    try {
      const response = await documentVersionService.getDocumentVersions(
        documentId,
        currentPage,
        20
      )
      setVersions(response.data.data)
      setTotalPages(response.data.total_pages)
      setPage(currentPage)
    } catch (err: any) {
      setError(err.message || '加载版本历史失败')
    } finally {
      setLoading(false)
    }
  }

  // 加载锁状态
  const loadLockStatus = async () => {
    try {
      const response = await documentVersionService.getDocumentLockStatus(documentId)
      setLockStatus(response.data.data)
    } catch (err) {
      console.error('Failed to load lock status:', err)
    }
  }

  useEffect(() => {
    loadVersions()
    loadLockStatus()
  }, [documentId])

  // 恢复版本
  const handleRestore = async (version: number) => {
    setRestoreDialog({ open: false, version: 0 })

    try {
      await documentVersionService.restoreDocumentVersion(documentId, version)
      await loadVersions()
      onVersionRestored?.(version)
    } catch (err: any) {
      setError(err.message || '恢复版本失败')
    }
  }

  // 删除版本
  const handleDelete = async (version: number) => {
    setDeleteDialog({ open: false, version: 0 })

    try {
      await documentVersionService.deleteDocumentVersion(documentId, version)
      await loadVersions()
    } catch (err: any) {
      setError(err.message || '删除版本失败')
    }
  }

  // 比较版本
  const handleCompare = async (fromVersion: number, toVersion: number) => {
    try {
      const response = await documentVersionService.compareDocumentVersions(
        documentId,
        fromVersion,
        toVersion
      )
      setCompareDialog({ open: true, comparison: response.data.data })
    } catch (err: any) {
      setError(err.message || '比较版本失败')
    }
  }

  const canEdit = lockStatus?.can_edit ?? !readOnly

  return (
    <Box sx={{ width: '100%', maxWidth: 800, mx: 'auto' }}>
      <Box sx={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', mb: 2 }}>
        <Typography variant="h6">
          {documentName ? `${documentName} - ` : ''}版本历史
        </Typography>
        {onClose && (
          <IconButton onClick={onClose} size="small">
            ×
          </IconButton>
        )}
      </Box>

      {error && (
        <Alert severity="error" sx={{ mb: 2 }} onClose={() => setError(null)}>
          {error}
        </Alert>
      )}

      {lockStatus && lockStatus.is_locked && !lockStatus.can_edit && (
        <Alert severity="warning" sx={{ mb: 2 }}>
          文档已被 {lockStatus.locked_by_name} 锁定，无法进行编辑操作
        </Alert>
      )}

      {loading ? (
        <Box sx={{ display: 'flex', justifyContent: 'center', py: 4 }}>
          <CircularProgress />
        </Box>
      ) : (
        <List>
          {versions.map((version, index) => (
            <React.Fragment key={version.id}>
              <ListItem
                sx={{
                  bgcolor: version.is_current ? 'action.selected' : 'background.paper',
                  mb: 1,
                  borderRadius: 1,
                  border: version.is_current ? 2 : 1,
                  borderColor: version.is_current ? 'primary.main' : 'divider',
                }}
              >
                <ListItemText
                  primary={
                    <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
                      <Typography variant="subtitle1" fontWeight="bold">
                        版本 {version.version}
                      </Typography>
                      {version.is_current && (
                        <Chip label="当前版本" size="small" color="primary" />
                      )}
                      <Chip
                        label={version.change_type === 'auto' ? '自动' : '手动'}
                        size="small"
                        variant="outlined"
                      />
                    </Box>
                  }
                  secondary={
                    <Box>
                      <Typography variant="body2" color="text.secondary">
                        {version.change_description || '无描述'}
                      </Typography>
                      <Typography variant="caption" color="text.secondary">
                        {version.created_by_name || '未知用户'} ·{' '}
                        {formatDistanceToNow(new Date(version.created_at), {
                          addSuffix: true,
                          locale: zhCN,
                        })}
                      </Typography>
                    </Box>
                  }
                />

                <ListItemSecondaryAction>
                  <Box sx={{ display: 'flex', gap: 0.5 }}>
                    {/* 查看按钮 */}
                    <Tooltip title="查看详情">
                      <IconButton
                        size="small"
                        onClick={() => window.open(`/api/documents/${documentId}/versions/${version.version}/download`)}
                      >
                        <VisibilityIcon fontSize="small" />
                      </IconButton>
                    </Tooltip>

                    {/* 下载按钮 */}
                    <Tooltip title="下载版本">
                      <IconButton
                        size="small"
                        onClick={() => window.open(`/api/documents/${documentId}/versions/${version.version}/download`)}
                      >
                        <DownloadIcon fontSize="small" />
                      </IconButton>
                    </Tooltip>

                    {/* 比较按钮 */}
                    {index < versions.length - 1 && canEdit && (
                      <Tooltip title="与上一版本比较">
                        <IconButton
                          size="small"
                          onClick={() => handleCompare(version.version, versions[index + 1].version)}
                        >
                          <CompareIcon fontSize="small" />
                        </IconButton>
                      </Tooltip>
                    )}

                    {/* 恢复按钮 */}
                    {!version.is_current && canEdit && (
                      <Tooltip title="恢复到此版本">
                        <IconButton
                          size="small"
                          color="primary"
                          onClick={() => setRestoreDialog({ open: true, version: version.version })}
                        >
                          <RestoreIcon fontSize="small" />
                        </IconButton>
                      </Tooltip>
                    )}

                    {/* 删除按钮 */}
                    {!version.is_current && versions.length > 1 && canEdit && (
                      <Tooltip title="删除版本">
                        <IconButton
                          size="small"
                          color="error"
                          onClick={() => setDeleteDialog({ open: true, version: version.version })}
                        >
                          <DeleteIcon fontSize="small" />
                        </IconButton>
                      </Tooltip>
                    )}
                  </Box>
                </ListItemSecondaryAction>
              </ListItem>
            </React.Fragment>
          ))}
        </List>
      )}

      {/* 分页 */}
      {totalPages > 1 && (
        <Box sx={{ display: 'flex', justifyContent: 'center', mt: 2, gap: 1 }}>
          <Button
            disabled={page <= 1}
            onClick={() => loadVersions(page - 1)}
            variant="outlined"
            size="small"
          >
            上一页
          </Button>
          <Typography variant="body2" sx={{ display: 'flex', alignItems: 'center' }}>
            第 {page} / {totalPages} 页
          </Typography>
          <Button
            disabled={page >= totalPages}
            onClick={() => loadVersions(page + 1)}
            variant="outlined"
            size="small"
          >
            下一页
          </Button>
        </Box>
      )}

      {/* 恢复确认对话框 */}
      <Dialog
        open={restoreDialog.open}
        onClose={() => setRestoreDialog({ open: false, version: 0 })}
      >
        <DialogTitle>确认恢复版本</DialogTitle>
        <DialogContent>
          <Typography>
            确定要恢复到版本 {restoreDialog.version} 吗？
          </Typography>
          <Typography variant="body2" color="text.secondary" sx={{ mt: 1 }}>
            当前版本将自动保存为备份。
          </Typography>
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setRestoreDialog({ open: false, version: 0 })}>
            取消
          </Button>
          <Button
            onClick={() => handleRestore(restoreDialog.version)}
            color="primary"
            variant="contained"
          >
            确认恢复
          </Button>
        </DialogActions>
      </Dialog>

      {/* 删除确认对话框 */}
      <Dialog
        open={deleteDialog.open}
        onClose={() => setDeleteDialog({ open: false, version: 0 })}
      >
        <DialogTitle>确认删除版本</DialogTitle>
        <DialogContent>
          <Typography>
            确定要删除版本 {deleteDialog.version} 吗？
          </Typography>
          <Typography variant="body2" color="error" sx={{ mt: 1 }}>
            此操作不可撤销！
          </Typography>
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setDeleteDialog({ open: false, version: 0 })}>
            取消
          </Button>
          <Button
            onClick={() => handleDelete(deleteDialog.version)}
            color="error"
            variant="contained"
          >
            确认删除
          </Button>
        </DialogActions>
      </Dialog>

      {/* 版本比较对话框 */}
      <CompareDialog
        open={compareDialog.open}
        comparison={compareDialog.comparison}
        onClose={() => setCompareDialog({ open: false, comparison: null })}
      />
    </Box>
  )
}

/**
 * 版本比较对话框组件
 */
const CompareDialog: React.FC<CompareDialogProps> = ({ open, comparison, onClose }) => {
  if (!comparison) return null

  return (
    <Dialog open={open} onClose={onClose} maxWidth="md" fullWidth>
      <DialogTitle>版本比较</DialogTitle>
      <DialogContent>
        <Box sx={{ mb: 2 }}>
          <Typography variant="subtitle2" color="text.secondary">
            从版本 {comparison.from_version.version} 到版本 {comparison.to_version.version}
          </Typography>
          <Typography variant="body2">{comparison.summary}</Typography>
        </Box>

        <Divider sx={{ my: 2 }} />

        <Typography variant="subtitle1" gutterBottom>
          变更详情
        </Typography>

        {comparison.changes.length === 0 ? (
          <Typography variant="body2" color="text.secondary">
            没有检测到变更
          </Typography>
        ) : (
          <List>
            {comparison.changes.map((change, index) => (
              <ListItem key={index} divider>
                <ListItemText
                  primary={
                    <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
                      <Chip
                        label={change.type === 'added' ? '新增' : change.type === 'removed' ? '删除' : '修改'}
                        size="small"
                        color={
                          change.type === 'added'
                            ? 'success'
                            : change.type === 'removed'
                            ? 'error'
                            : 'info'
                        }
                      />
                      <Typography variant="body2">{change.path}</Typography>
                    </Box>
                  }
                  secondary={change.description}
                />
              </ListItem>
            ))}
          </List>
        )}
      </DialogContent>
      <DialogActions>
        <Button onClick={onClose}>关闭</Button>
      </DialogActions>
    </Dialog>
  )
}

export default VersionHistory
