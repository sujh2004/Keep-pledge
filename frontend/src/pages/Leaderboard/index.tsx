import { useState, useEffect } from 'react'
import { Card, Tabs, Table, Avatar, Tag, Space, Typography, Spin } from 'antd'
import { CrownOutlined, UserOutlined, ThunderboltOutlined } from '@ant-design/icons'
import { leaderboardApi } from '../../api'
import { useAuthStore } from '../../store/auth'
import type { UserSummary } from '../../types'

export default function LeaderboardPage() {
  const [weekly, setWeekly] = useState<UserSummary[]>([])
  const [friendsBoard, setFriendsBoard] = useState<UserSummary[]>([])
  const [loading, setLoading] = useState(true)
  const currentUser = useAuthStore((s) => s.user)

  useEffect(() => {
    setLoading(true)
    Promise.all([leaderboardApi.weekly(), leaderboardApi.friends()])
      .then(([w, f]) => { setWeekly(w); setFriendsBoard(f) })
      .catch(() => {})
      .finally(() => setLoading(false))
  }, [])

  const columns = [
    {
      title: '排名',
      key: 'rank',
      width: 80,
      render: (_: unknown, __: unknown, index: number) => {
        if (index === 0) return <CrownOutlined style={{ color: '#faad14', fontSize: 20 }} />
        if (index === 1) return <CrownOutlined style={{ color: '#bfbfbf', fontSize: 18 }} />
        if (index === 2) return <CrownOutlined style={{ color: '#d48806', fontSize: 16 }} />
        return index + 1
      },
    },
    {
      title: '用户',
      key: 'user',
      render: (_: unknown, record: UserSummary) => (
        <Space>
          <Avatar src={record.avatar || undefined} icon={<UserOutlined />} />
          <span>{record.username}</span>
          {record.id === currentUser?.id && <Tag color="purple">我</Tag>}
        </Space>
      ),
    },
    {
      title: '等级',
      key: 'level',
      render: (_: unknown, record: UserSummary) => (
        <Tag color="blue"><ThunderboltOutlined /> Lv.{record.level}</Tag>
      ),
    },
    {
      title: 'XP',
      dataIndex: 'xp',
      key: 'xp',
      sorter: (a: UserSummary, b: UserSummary) => b.xp - a.xp,
    },
    {
      title: '信用分',
      dataIndex: 'credit_score',
      key: 'credit_score',
    },
  ]

  return (
    <Card title={<Space><CrownOutlined /> 排行榜</Space>}>
      <Spin spinning={loading}>
        <Tabs
          items={[
            {
              key: 'weekly',
              label: '综合周榜',
              children: (
                <Table
                  dataSource={weekly}
                  columns={columns}
                  rowKey="id"
                  pagination={false}
                  rowClassName={(record) => record.id === currentUser?.id ? 'bg-indigo-50' : ''}
                />
              ),
            },
            {
              key: 'friends',
              label: '好友榜',
              children: friendsBoard.length === 0 ? (
                <Typography.Text type="secondary">添加好友后即可查看好友排行</Typography.Text>
              ) : (
                <Table
                  dataSource={friendsBoard}
                  columns={columns}
                  rowKey="id"
                  pagination={false}
                  rowClassName={(record) => record.id === currentUser?.id ? 'bg-indigo-50' : ''}
                />
              ),
            },
          ]}
        />
      </Spin>
    </Card>
  )
}
