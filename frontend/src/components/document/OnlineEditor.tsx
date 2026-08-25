import React, { useEffect, useRef, useState } from 'react'
import {
  Box,
  Dialog,
  DialogContent,
  DialogTitle,
  IconButton,
  Toolbar,
  Tooltip,
  Typography,
  Alert,
  Chip,
  CircularProgress,
} from '@mui/material'
import {
  Close as CloseIcon,
  Lock as LockIcon,
  LockOpen as LockOpenIcon,
  History as HistoryIcon,
  Refresh as RefreshIcon,
} from '@mui/icons-material'
import documentVersionService, { LockStatus, OnlyOfficeConfig } from '@/services/documentVersion'
import { getOnlyOfficeOrigin, isOnlyOfficeEnabled } from '@/config/features'
import {
  buildOnlyOfficeScriptUrl,
  encodeJsonForScript,
  getOnlyOfficeAllowedOrigins,
} from '@/config/onlyoffice'

// ============================================================================
// 类型定义
// ============================================================================

interface OnlineEditorProps {
  documentId: number
  documentName?: string
  open: boolean
  onClose: () => void
  mode?: 'edit' | 'view'
  onSave?: () => void
}

interface OnlyOfficeAPI {
  processSaveResult: (result: boolean) => void
  processRightsChange: (enabled: boolean) => void
  setActionLink: (data: any) => void
}

const onlyOfficeEnabled = isOnlyOfficeEnabled()
const onlyOfficeOrigin = getOnlyOfficeOrigin()
const onlyOfficeDisabledMessage = 'OnlyOffice 未启用，在线编辑和转换已关闭。'

declare global {
  interface Window {
    DocEditor?: any
  }
}

// ============================================================================
// 组件
// ============================================================================

/**
 * OnlyOffice 在线编辑器组件
 */
export const OnlineEditor: React.FC<OnlineEditorProps> = ({
  documentId,
  documentName,
  open,
  onClose,
  mode = 'edit',
  onSave,
}) => {
  const iframeRef = useRef<HTMLIFrameElement>(null)
  const [config, setConfig] = useState<OnlyOfficeConfig | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [lockStatus, setLockStatus] = useState<LockStatus | null>(null)
  const [showVersionHistory, setShowVersionHistory] = useState(false)
  const [connectionStatus, setConnectionStatus] = useState<'connected' | 'disconnected' | 'saving'>(
    'connected',
  )

  // 加载编辑器配置
  const loadEditorConfig = async () => {
    if (!onlyOfficeEnabled) {
      setLoading(false)
      return
    }
    setLoading(true)
    setError(null)

    try {
      const response = await documentVersionService.openOnlyOfficeEditor(documentId, mode)
      setConfig(response.data.data)
    } catch (err: any) {
      if (err.response?.status === 423) {
        // 文档被锁定
        setError(err.response.data?.error || '文档已被其他用户锁定')
      } else {
        setError(err.message || '加载编辑器失败')
      }
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

  // 续期锁
  const renewLock = async () => {
    try {
      await documentVersionService.renewDocumentLock(documentId)
      await loadLockStatus()
    } catch (err) {
      console.error('Failed to renew lock:', err)
    }
  }

  // 释放锁
  const releaseLock = async () => {
    try {
      await documentVersionService.releaseDocumentLock(documentId)
    } catch (err) {
      console.error('Failed to release lock:', err)
    }
  }

  // 处理来自 OnlyOffice 的消息
  useEffect(() => {
    if (!onlyOfficeEnabled) {
      return
    }
    const handleMessage = (event: MessageEvent) => {
      // 验证消息来源
      const allowedOrigins = new Set(
        getOnlyOfficeAllowedOrigins(onlyOfficeOrigin || null, window.location.origin),
      )
      const isSandboxedEditorMessage =
        event.source === iframeRef.current?.contentWindow && event.origin === 'null'
      if (!isSandboxedEditorMessage && !allowedOrigins.has(event.origin)) {
        return
      }

      // 处理不同类型的事件
      switch (event.data) {
        case 'onSave':
          setConnectionStatus('saving')
          break
        case 'onDocumentStateChange':
          setConnectionStatus('connected')
          break
        case 'onDocumentReady':
          setConnectionStatus('connected')
          setLoading(false)
          break
      }
    }

    window.addEventListener('message', handleMessage)
    return () => window.removeEventListener('message', handleMessage)
  }, [])

  // 定时续期锁
  useEffect(() => {
    if (!onlyOfficeEnabled) {
      return
    }
    if (!open || !lockStatus?.is_locked || !lockStatus.can_edit) {
      return
    }

    // 每 5 分钟续期一次
    const interval = setInterval(renewLock, 5 * 60 * 1000)
    return () => clearInterval(interval)
  }, [open, lockStatus])

  // 初始化时加载配置
  useEffect(() => {
    if (!onlyOfficeEnabled) {
      setConfig(null)
      setLockStatus(null)
      setShowVersionHistory(false)
      setLoading(false)
      return
    }
    if (open && documentId) {
      loadEditorConfig()
      loadLockStatus()
    }
  }, [open, documentId, mode])

  // 关闭时释放锁
  useEffect(() => {
    return () => {
      if (onlyOfficeEnabled && mode === 'edit') {
        releaseLock()
      }
    }
  }, [])

  const handleRefresh = () => {
    if (!onlyOfficeEnabled) {
      return
    }
    loadEditorConfig()
    loadLockStatus()
  }

  return (
    <Dialog
      open={open}
      onClose={onClose}
      maxWidth='xl'
      fullWidth
      PaperProps={{
        sx: {
          height: '90vh',
          maxHeight: '90vh',
        },
      }}
    >
      <DialogTitle sx={{ pb: 1 }}>
        <Box sx={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
          <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
            <Typography variant='h6' component='span'>
              {documentName || '文档编辑'}
            </Typography>
            {lockStatus?.is_locked && (
              <Chip
                icon={lockStatus.can_edit ? <LockOpenIcon /> : <LockIcon />}
                label={
                  lockStatus.can_edit ? '已锁定（您）' : `已锁定（${lockStatus.locked_by_name}）`
                }
                size='small'
                color={lockStatus.can_edit ? 'success' : 'warning'}
              />
            )}
            {connectionStatus === 'saving' && <Chip label='保存中...' size='small' color='info' />}
            {connectionStatus === 'disconnected' && (
              <Chip label='连接断开' size='small' color='error' />
            )}
          </Box>

          <Box sx={{ display: 'flex', alignItems: 'center', gap: 0.5 }}>
            <Tooltip title='刷新状态'>
              <span>
                <IconButton onClick={handleRefresh} size='small' disabled={!onlyOfficeEnabled}>
                  <RefreshIcon />
                </IconButton>
              </span>
            </Tooltip>

            {onlyOfficeEnabled && mode === 'edit' && (
              <Tooltip title='版本历史'>
                <IconButton
                  onClick={() => setShowVersionHistory(!showVersionHistory)}
                  size='small'
                  color={showVersionHistory ? 'primary' : 'default'}
                >
                  <HistoryIcon />
                </IconButton>
              </Tooltip>
            )}

            <Tooltip title='关闭'>
              <IconButton onClick={onClose} size='small'>
                <CloseIcon />
              </IconButton>
            </Tooltip>
          </Box>
        </Box>
      </DialogTitle>

      <DialogContent sx={{ p: 0, height: 'calc(100% - 72px)', display: 'flex' }}>
        <Box sx={{ width: '100%', height: '100%', position: 'relative' }}>
          {error && (
            <Alert
              severity={onlyOfficeEnabled ? 'error' : 'info'}
              sx={{ position: 'absolute', top: 8, left: 8, right: 8, zIndex: 1 }}
            >
              {error}
            </Alert>
          )}

          {!onlyOfficeEnabled && (
            <Box
              sx={{
                height: '100%',
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'center',
                p: 4,
              }}
            >
              <Alert severity='info' sx={{ width: '100%' }}>
                {onlyOfficeDisabledMessage}
              </Alert>
            </Box>
          )}

          {onlyOfficeEnabled && loading && !error && (
            <Box
              sx={{
                position: 'absolute',
                top: 0,
                left: 0,
                right: 0,
                bottom: 0,
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'center',
                flexDirection: 'column',
                gap: 2,
              }}
            >
              <CircularProgress />
              <Typography variant='body2' color='text.secondary'>
                正在加载编辑器...
              </Typography>
            </Box>
          )}

          {onlyOfficeEnabled && config && !loading && (
            <iframe
              ref={iframeRef}
              srcDoc={generateEditorHTML(config, onlyOfficeOrigin, window.location.origin)}
              sandbox='allow-scripts allow-forms'
              style={{
                width: '100%',
                height: '100%',
                border: 'none',
              }}
              title='OnlyOffice Editor'
            />
          )}

          {/* 版本历史侧边栏 */}
          {onlyOfficeEnabled && showVersionHistory && (
            <Box
              sx={{
                position: 'absolute',
                top: 0,
                right: 0,
                width: 400,
                height: '100%',
                bgcolor: 'background.paper',
                borderLeft: 1,
                borderColor: 'divider',
                zIndex: 2,
                overflow: 'auto',
              }}
            >
              <VersionHistorySidebar
                documentId={documentId}
                documentName={documentName}
                onClose={() => setShowVersionHistory(false)}
                onVersionRestored={() => {
                  setShowVersionHistory(false)
                  handleRefresh()
                }}
              />
            </Box>
          )}
        </Box>
      </DialogContent>
    </Dialog>
  )
}

// ============================================================================
// 子组件
// ============================================================================

interface VersionHistorySidebarProps {
  documentId: number
  documentName?: string
  onClose: () => void
  onVersionRestored?: (version: number) => void
}

const VersionHistorySidebar: React.FC<VersionHistorySidebarProps> = ({
  documentId,
  documentName,
  onClose,
  onVersionRestored,
}) => {
  const [versions, setVersions] = useState<any[]>([])
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    const loadVersions = async () => {
      setLoading(true)
      try {
        const response = await documentVersionService.getDocumentVersions(documentId, 1, 50)
        setVersions(response.data.data)
      } catch (err) {
        console.error('Failed to load versions:', err)
      } finally {
        setLoading(false)
      }
    }

    loadVersions()
  }, [documentId])

  return (
    <Box>
      <Toolbar variant='dense'>
        <Typography variant='subtitle1' sx={{ flex: 1 }}>
          版本历史
        </Typography>
        <IconButton onClick={onClose} size='small'>
          ×
        </IconButton>
      </Toolbar>

      {loading ? (
        <Box sx={{ display: 'flex', justifyContent: 'center', py: 4 }}>
          <CircularProgress size={24} />
        </Box>
      ) : (
        <Box sx={{ px: 2 }}>
          {versions.map((version) => (
            <Box
              key={version.id}
              sx={{
                py: 1,
                borderBottom: 1,
                borderColor: 'divider',
                '&:last-child': { borderBottom: 'none' },
              }}
            >
              <Typography variant='subtitle2'>
                版本 {version.version}
                {version.is_current && ' (当前)'}
              </Typography>
              <Typography variant='body2' color='text.secondary'>
                {version.change_description || '-'}
              </Typography>
              <Typography variant='caption' color='text.secondary'>
                {version.created_by_name} · {new Date(version.created_at).toLocaleString()}
              </Typography>
            </Box>
          ))}
        </Box>
      )}
    </Box>
  )
}

// ============================================================================
// 工具函数
// ============================================================================

/**
 * 生成 OnlyOffice 编辑器 HTML
 */
function generateEditorHTML(
  config: OnlyOfficeConfig,
  onlyOfficeOrigin: string,
  parentOrigin: string,
): string {
  const scriptUrl = buildOnlyOfficeScriptUrl(onlyOfficeOrigin)
  return `
<!DOCTYPE html>
<html>
<head>
  <meta charset="utf-8" />
  <meta name="viewport" content="width=device-width, initial-scale=1.0" />
  <title>OnlyOffice Editor</title>
  <style>
    html, body {
      margin: 0;
      padding: 0;
      width: 100%;
      height: 100%;
      overflow: hidden;
    }
    #editor {
      width: 100%;
      height: 100%;
    }
  </style>
</head>
<body>
  <div id="editor"></div>

  <script src="${scriptUrl}"></script>
  <script>
    const docEditor = new DocsAPI.DocEditor("editor", {
      ...${encodeJsonForScript(config)},
      events: {
        onReady: function() {
          window.parent.postMessage('onDocumentReady', ${JSON.stringify(parentOrigin)});
        },
        onDocumentStateChange: function() {
          window.parent.postMessage('onDocumentStateChange', ${JSON.stringify(parentOrigin)});
        },
        onSave: function(event) {
          window.parent.postMessage('onSave', ${JSON.stringify(parentOrigin)});
        },
        onError: function(event) {
          console.error('OnlyOffice error:', event);
          window.parent.postMessage('onError', ${JSON.stringify(parentOrigin)});
        }
      }
    });
  </script>
</body>
</html>
  `
}

export default OnlineEditor
