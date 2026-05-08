import { useState, useEffect } from 'react'
import { Card, Row, Col, Typography, Tag, Spin, Space, Tabs } from 'antd'
import {
  StarOutlined, CheckCircleOutlined, LockOutlined,
  FireOutlined, TrophyOutlined, TeamOutlined, QuestionCircleOutlined,
} from '@ant-design/icons'
import { achievementApi } from '../../api'
import type { AchievementResponse } from '../../types'

const categoryIcons: Record<string, React.ReactNode> = {
  checkin: <FireOutlined />,
  challenge: <TrophyOutlined />,
  social: <TeamOutlined />,
  hidden: <QuestionCircleOutlined />,
}

const categoryLabels: Record<string, string> = {
  checkin: '打卡类',
  challenge: '挑战类',
  social: '社交类',
  hidden: '隐藏类',
}

export default function AchievementsPage() {
  const [achievements, setAchievements] = useState<AchievementResponse[]>([])
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    achievementApi.all()
      .then(setAchievements)
      .catch(() => {})
      .finally(() => setLoading(false))
  }, [])

  const unlocked = achievements.filter((a) => a.unlocked).length

  const grouped = achievements.reduce<Record<string, AchievementResponse[]>>((acc, a) => {
    const cat = a.category
    if (!acc[cat]) acc[cat] = []
    acc[cat].push(a)
    return acc
  }, {})

  return (
    <Card
      title={
        <Space>
          <StarOutlined />
          <span>成就墙</span>
          <Tag color="purple">{unlocked}/{achievements.length} 已解锁</Tag>
        </Space>
      }
    >
      <Spin spinning={loading}>
        <Tabs
          items={Object.entries(grouped).map(([cat, items]) => ({
            key: cat,
            label: (
              <Space>
                {categoryIcons[cat]}
                {categoryLabels[cat] || cat}
                <Tag>{items.filter((i) => i.unlocked).length}/{items.length}</Tag>
              </Space>
            ),
            children: (
              <Row gutter={[16, 16]}>
                {items.map((a) => (
                  <Col xs={24} sm={12} md={8} lg={6} key={a.id}>
                    <Card
                      hoverable={a.unlocked}
                      className={a.unlocked ? 'border-2 border-yellow-300' : 'opacity-50'}
                    >
                      <div className="text-center">
                        <div className="text-4xl mb-3">
                          {a.is_hidden && !a.unlocked ? (
                            <LockOutlined style={{ color: '#d9d9d9' }} />
                          ) : (
                            <StarOutlined style={{ color: a.unlocked ? '#faad14' : '#d9d9d9' }} />
                          )}
                        </div>
                        <Typography.Title level={5}>
                          {a.is_hidden && !a.unlocked ? '???' : a.name}
                        </Typography.Title>
                        <Typography.Text type="secondary" className="text-xs">
                          {a.is_hidden && !a.unlocked ? '完成特殊条件解锁' : a.description}
                        </Typography.Text>
                        <div className="mt-2">
                          {a.unlocked ? (
                            <Tag icon={<CheckCircleOutlined />} color="success">
                              已解锁 {a.unlocked_at?.slice(0, 10)}
                            </Tag>
                          ) : (
                            <Tag icon={<LockOutlined />}>未解锁</Tag>
                          )}
                        </div>
                      </div>
                    </Card>
                  </Col>
                ))}
              </Row>
            ),
          }))}
        />
      </Spin>
    </Card>
  )
}
