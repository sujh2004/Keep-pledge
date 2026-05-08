import { useState, useEffect } from 'react'
import { Badge } from 'antd'
import { BellOutlined } from '@ant-design/icons'
import { notificationApi } from '../../api'

export default function NotificationBell({ onClick }: { onClick?: () => void }) {
  const [count, setCount] = useState(0)

  useEffect(() => {
    const poll = () => {
      notificationApi.unreadCount().then((r) => setCount(r.count)).catch(() => {})
    }
    poll()
    const timer = setInterval(poll, 30000)
    return () => clearInterval(timer)
  }, [])

  return (
    <Badge count={count} size="small">
      <BellOutlined style={{ fontSize: 20, cursor: 'pointer' }} onClick={onClick} />
    </Badge>
  )
}
