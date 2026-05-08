import { useState, useEffect } from 'react'
import { useNavigate } from 'react-router-dom'
import {
  Card, Row, Col, Tag, Space, Typography, Button, Spin, Empty, message, Tabs,
} from 'antd'
import { CompassOutlined, UserOutlined, FireOutlined } from '@ant-design/icons'
import { challengeApi } from '../../api'
import { CATEGORY_LABELS } from '../../types'
import { categoryIcon } from '../../utils'
import type { ChallengeResponse, ChallengeCategory } from '../../types'
import { statusLabel } from '../../utils'

const categories: { key: string; label: string }[] = [
  { key: '', label: '全部' },
  ...Object.entries(CATEGORY_LABELS).map(([k, v]) => ({ key: k, label: v })),
]

export default function ExplorePage() {
  const [challenges, setChallenges] = useState<ChallengeResponse[]>([])
  const [loading, setLoading] = useState(true)
  const [category, setCategory] = useState('')
  const navigate = useNavigate()

  useEffect(() => {
    setLoading(true)
    challengeApi.explore(category || undefined)
      .then(setChallenges)
      .catch(() => {})
      .finally(() => setLoading(false))
  }, [category])

  const handleJoin = async (id: number) => {
    try {
      await challengeApi.join(id)
      message.success('加入成功')
      navigate(`/challenge/${id}`)
    } catch (err: unknown) {
      message.error(err instanceof Error ? err.message : '加入失败')
    }
  }

  return (
    <Card title={<Space><CompassOutlined /> 挑战广场</Space>}>
      <Tabs
        activeKey={category}
        onChange={setCategory}
        items={categories.map((c) => ({ key: c.key, label: c.label }))}
        className="mb-4"
      />
      <Spin spinning={loading}>
        {challenges.length === 0 ? (
          <Empty description="暂无公开挑战" />
        ) : (
          <Row gutter={[16, 16]}>
            {challenges.map((ch) => {
              const status = statusLabel(ch.status)
              return (
                <Col xs={24} sm={12} lg={8} key={ch.id}>
                  <Card hoverable>
                    <div className="flex items-center justify-between mb-2">
                      <Space>
                        <span>{categoryIcon(ch.category)}</span>
                        <Typography.Text strong>{ch.title}</Typography.Text>
                      </Space>
                      <Tag color={status.color}>{status.text}</Tag>
                    </div>
                    <Typography.Paragraph type="secondary" ellipsis={{ rows: 2 }}>
                      {ch.description || ch.pledge}
                    </Typography.Paragraph>
                    <div className="flex justify-between items-center text-xs text-gray-400">
                      <Space>
                        <span>{CATEGORY_LABELS[ch.category as ChallengeCategory]}</span>
                        <span>{ch.target_days}天</span>
                        <span><UserOutlined /> {ch.participant_count}</span>
                      </Space>
                      <Button
                        type="primary"
                        size="small"
                        onClick={(e) => { e.stopPropagation(); handleJoin(ch.id) }}
                      >
                        加入挑战
                      </Button>
                    </div>
                  </Card>
                </Col>
              )
            })}
          </Row>
        )}
      </Spin>
    </Card>
  )
}
