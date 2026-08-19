import { ReactNode, useEffect, useState } from 'react'
import { createPortal } from 'react-dom'
import { X } from 'lucide-react'
import { Button } from './Button'

interface ModalProps {
  open: boolean
  onClose: () => void
  title?: string
  children: ReactNode
  footer?: ReactNode
  width?: string
  closable?: boolean
}

export function Modal({
  open,
  onClose,
  title,
  children,
  footer,
  width = '500px',
  closable = true,
}: ModalProps) {
  useEffect(() => {
    if (open) {
      document.body.style.overflow = 'hidden'
    } else {
      document.body.style.overflow = ''
    }
    return () => {
      document.body.style.overflow = ''
    }
  }, [open])

  if (!open) return null

  return createPortal(
    <div className="fixed inset-0 z-[9990] flex items-center justify-center">
      <div
        className="absolute inset-0 bg-black/50 backdrop-blur-sm animate-fade-in"
        onClick={closable ? onClose : undefined}
      />

      <div
        className="relative bg-[var(--color-bg-elevated)] rounded-xl shadow-2xl animate-scale-in max-h-[90vh] w-full flex flex-col"
        style={{ width, maxWidth: '90vw' }}
        onClick={(e) => e.stopPropagation()}
      >
        {(title || closable) && (
          <div className="flex items-center justify-between px-6 py-4 border-b border-[var(--color-border)] flex-shrink-0">
            {title && (
              <h3 className="text-lg font-semibold text-[var(--color-text-primary)]">
                {title}
              </h3>
            )}
            {closable && (
              <button
                onClick={onClose}
                className="p-1.5 rounded-lg text-[var(--color-text-muted)] hover:text-[var(--color-text-primary)] hover:bg-[var(--color-bg-muted)] transition-colors ml-auto"
              >
                <X className="w-5 h-5" />
              </button>
            )}
          </div>
        )}

        <div className="px-6 py-4 overflow-y-auto flex-1 min-h-0">
          {children}
        </div>

        {footer && (
          <div className="flex items-center justify-end gap-3 px-6 py-4 border-t border-[var(--color-border)] flex-shrink-0">
            {footer}
          </div>
        )}
      </div>
    </div>,
    document.body,
  )
}

// 确认对话框
interface ConfirmModalProps {
  open: boolean
  onClose: () => void
  onConfirm: () => boolean | void | Promise<boolean | void>
  title?: string
  content: ReactNode
  confirmText?: string
  cancelText?: string
  danger?: boolean
}

export function ConfirmModal({
  open,
  onClose,
  onConfirm,
  title = '确认',
  content,
  confirmText = '确定',
  cancelText = '取消',
  danger = false,
}: ConfirmModalProps) {
  const [confirming, setConfirming] = useState(false)
  const [confirmError, setConfirmError] = useState('')

  useEffect(() => {
    if (!open) {
      setConfirming(false)
      setConfirmError('')
    }
  }, [open])

  const handleClose = () => {
    if (confirming) return
    setConfirmError('')
    onClose()
  }

  const handleConfirm = async () => {
    if (confirming) return
    setConfirming(true)
    setConfirmError('')
    try {
      const confirmed = await onConfirm()
      if (confirmed !== false) {
        onClose()
      }
    } catch (error: any) {
      setConfirmError(error?.message || '操作失败，请重试')
    } finally {
      setConfirming(false)
    }
  }

  return (
    <Modal
      open={open}
      onClose={handleClose}
      title={title}
      width="400px"
      closable={!confirming}
      footer={
        <>
          <Button variant="secondary" onClick={handleClose} disabled={confirming}>
            {cancelText}
          </Button>
          <Button
            variant={danger ? 'danger' : 'primary'}
            onClick={handleConfirm}
            loading={confirming}
          >
            {confirmText}
          </Button>
        </>
      }
    >
      <div className="space-y-3">
        <div className="text-[var(--color-text-secondary)]">{content}</div>
        {confirmError && (
          <div role="alert" className="text-sm text-[var(--color-error)]">{confirmError}</div>
        )}
      </div>
    </Modal>
  )
}
