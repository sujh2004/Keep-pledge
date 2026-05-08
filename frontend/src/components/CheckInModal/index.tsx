import { useState } from 'react'
import { Modal, Input, Upload, message, Space, Typography } from 'antd'
import { PlusOutlined } from '@ant-design/icons'
import { checkinApi } from '../../api'
import type { SubmitCheckInResponse } from '../../types'

interface Props {
  challengeId: number
  open: boolean
  onClose: () => void
  onSuccess: (result: SubmitCheckInResponse) => void
}

export default function CheckInModal({ challengeId, open, onClose, onSuccess }: Props) {
  const [content, setContent] = useState('')
  const [file, setFile] = useState<File | null>(null)
  const [loading, setLoading] = useState(false)

  const handleSubmit = async () => {
    if (!content.trim() && !file) {
      message.warning('请输入打卡内容或上传图片')
      return
    }
    setLoading(true)
    try {
      const result = await checkinApi.submit(challengeId, content, file || undefined)
      message.success(`打卡成功! +${result.xp_earned} XP`)
      if (result.achievements_unlocked?.length) {
        result.achievements_unlocked.forEach((a) => {
          message.success(`解锁成就: ${a.name}!`)
        })
      }
      setContent('')
      setFile(null)
      onSuccess(result)
      onClose()
    } catch (err: unknown) {
      message.error(err instanceof Error ? err.message : '打卡失败')
    } finally {
      setLoading(false)
    }
  }

  return (
    <Modal
      title="今日打卡"
      open={open}
      onOk={handleSubmit}
      onCancel={onClose}
      confirmLoading={loading}
      okText="提交打卡"
      cancelText="取消"
      destroyOnClose
    >
      <Space direction="vertical" className="w-full" size="middle">
        <div>
          <Typography.Text type="secondary">打卡内容</Typography.Text>
          <Input.TextArea
            rows={4}
            placeholder="记录今天的打卡心得..."
            value={content}
            onChange={(e) => setContent(e.target.value)}
            maxLength={500}
            showCount
          />
        </div>
        <div>
          <Typography.Text type="secondary">上传图片（可选）</Typography.Text>
          <Upload
            listType="picture-card"
            maxCount={1}
            beforeUpload={(f) => {
              setFile(f)
              return false
            }}
            onRemove={() => setFile(null)}
            accept="image/jpeg,image/png,image/webp"
          >
            {!file && (
              <div>
                <PlusOutlined />
                <div style={{ marginTop: 8 }}>上传</div>
              </div>
            )}
          </Upload>
        </div>
      </Space>
    </Modal>
  )
}
