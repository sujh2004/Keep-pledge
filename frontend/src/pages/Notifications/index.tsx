import { useState, useEffect } from 'react'
import { useNavigate } from 'react-router-dom'
import { Card, List, Tag, Typography, Button, Space, Spin, Empty, message } from 'antd'
import {
  BellOutlined, CheckCircleOutlined, UserAddOutlined, FireOutlined,
  TrophyOutlined, LikeOutlined, CommentOutlined, WarningOutlined,
  TeamOutlined, NotificationOutlined,
} from '@ant-design/icons'
import { notificationApi } from '../../api'
import type { Notification } from '../../types'

const typeConfig: Record<string, { icon: React.ReactNode; color: string }> = {
  invite: { icon: <TeamOutlined />, color: 'purple' },
  checkin_remind: { icon: <FireOutlined />, color: 'orange' },
  streak_warn: { icon: <WarningOutlined />, color: 'red' },
  achievement: { icon: <TrophyOutlined />, color: 'gold' },
  witness: { icon: <CheckCircleOutlined />, color: 'cyan' },
  friend_request: { icon: <UserAddOutlined />, color: 'blue' },
  like: { icon: <LikeOutlined />, color: 'pink' },
  comment: { icon: <CommentOutlined />, color: 'green' },
  system: { icon: <NotificationOutlined />, color: 'default' },
}

export default function NotificationsPage() {
  const [notifications, setNotifications] = useState<Notification[]>([])
  const [loading, setLoading] = useState(true)
  const navigate = useNavigate()

  const load = () => {
    setLoading(true)
    notificationApi.list()
      .then(setNotifications)
      .catch(() => {})
      .finally(() => setLoading(false))
  }

  useEffect(() => { load() }, [])

  const handleMarkRead = async (id: number) => {
    try {
      await notificationApi.markRead(id)
      setNotifications((prev) => prev.map((n) => (n.id === id ? { ...n, is_read: true } : n)))
    } catch {
      // ignore
    }
  }

  const handleMarkAllRead = async () => {
    try {
      await notificationApi.markAllRead()
      setNotifications((prev) => prev.map((n) => ({ ...n, is_read: true })))
      message.success('已全部标记为已读')
    } catch {
      message.error('操作失败')
    }
  }

  const handleClick = (n: Notification) => {
    handleMarkRead(n.id)
    if (n.related_type === 'challenge' && n.related_id) {
      navigate(`/challenge/${n.related_id}`)
    } else if (n.related_type === 'user' && n.related_id) {
      navigate('/friends')
    } else if (n.related_type === 'achievement') {
      navigate('/achievements')
    }
  }

  const unreadCount = notifications.filter((n) => !n.is_read).length

  return (
    <Card
      title={
        <Space>
          <BellOutlined />
          <span>通知中心</span>
          {unreadCount > 0 && <Tag color="red">{unreadCount} 未读</Tag>}
        </Space>
      }
      extra={
        unreadCount > 0 && (
          <Button size="small" onClick={handleMarkAllRead}>
            全部已读
          </Button>
        )
      }
    >
      <Spin spinning={loading}>
        {notifications.length === 0 ? (
          <Empty description="暂无通知" />
        ) : (
          <List
            dataSource={notifications}
            renderItem={(n) => {
              const config = typeConfig[n.type] || typeConfig.system
              return (
                <List.Item
                  className={`cursor-pointer ${n.is_read ? '' : 'bg-blue-50'}`}
                  onClick={() => handleClick(n)}
                  style={{ borderRadius: 8, marginBottom: 4, padding: '12px 16px' }}
                >
                  <List.Item.Meta
                    avatar={
                      <div
                        className="flex items-center justify-center rounded-full"
                        style={{ width: 40, height: 40, background: '#f5f5f5' }}
                      >
                        {config.icon}
                      </div>
                    }
                    title={
                      <Space>
                        <Typography.Text strong={!n.is_read}>{n.title}</Typography.Text>
                        {!n.is_read && <Tag color="blue" className="text-xs">新</Tag>}
                      </Space>
                    }
                    description={
                      <div>
                        <Typography.Text type="secondary">{n.content}</Typography.Text>
                        <br />
                        <Typography.Text type="secondary" className="text-xs">
                          {new Date(n.created_at).toLocaleString('zh-CN')}
                        </Typography.Text>
                      </div>
                    }
                  />
                </List.Item>
              )
            }}
          />
        )}
      </Spin>
    </Card>
  )
}
