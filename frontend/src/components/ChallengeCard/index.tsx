import { useNavigate } from 'react-router-dom'
import { Card, Progress, Tag, Typography, Space } from 'antd'
import { FireOutlined } from '@ant-design/icons'
import type { ChallengeResponse } from '../../types'
import { CATEGORY_LABELS } from '../../types'
import { categoryIcon } from '../../utils'

export default function ChallengeCard({ challenge }: { challenge: ChallengeResponse }) {
  const navigate = useNavigate()
  const status = statusLabel(challenge.status)

  return (
    <Card
      hoverable
      onClick={() => navigate(`/challenge/${challenge.id}`)}
      className="mb-4"
    >
      <Space direction="vertical" className="w-full" size="small">
        <div className="flex items-center justify-between">
          <Space>
            <span>{categoryIcon(challenge.category)}</span>
            <Typography.Text strong>{challenge.title}</Typography.Text>
          </Space>
          <Tag color={status.color}>{status.text}</Tag>
        </div>
        <Typography.Text type="secondary" ellipsis>
          {challenge.pledge}
        </Typography.Text>
        <Progress
          percent={challenge.progress}
          size="small"
          strokeColor="#6366f1"
        />
        <div className="flex justify-between text-xs text-gray-400">
          <Space>
            <span>{CATEGORY_LABELS[challenge.category as keyof typeof CATEGORY_LABELS] || challenge.category}</span>
            <span>·</span>
            <span>{challenge.target_days}天</span>
          </Space>
          <Space>
            <FireOutlined />
            <span>{challenge.checkin_count}/{challenge.target_days}</span>
            <span>·</span>
            <span>{challenge.participant_count}人参与</span>
          </Space>
        </div>
      </Space>
    </Card>
  )
}

function statusLabel(status: string): { text: string; color: string } {
  const map: Record<string, { text: string; color: string }> = {
    pending: { text: '待开始', color: 'blue' },
    active: { text: '进行中', color: 'green' },
    completed: { text: '已完成', color: 'gold' },
    failed: { text: '已失败', color: 'red' },
    cancelled: { text: '已取消', color: 'default' },
  }
  return map[status] || { text: status, color: 'default' }
}
