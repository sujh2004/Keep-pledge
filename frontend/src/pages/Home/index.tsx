import { useState, useEffect } from 'react'
import { useNavigate } from 'react-router-dom'
import { Card, Row, Col, Typography, Progress, Space, Button, Tag, Tooltip, Statistic, Empty, Spin, message } from 'antd'
import {
  FireOutlined,
  PlusOutlined,
  ThunderboltOutlined,
  CheckCircleOutlined,
  SafetyCertificateOutlined,
} from '@ant-design/icons'
import { useAuthStore } from '../../store/auth'
import { challengeApi, checkinApi, userApi } from '../../api'
import { creditLabel, levelXp, categoryIcon } from '../../utils'
import type { ChallengeResponse, HeatmapEntry } from '../../types'
import ChallengeCard from '../../components/ChallengeCard'
import CheckInModal from '../../components/CheckInModal'

export default function HomePage() {
  const user = useAuthStore((s) => s.user)
  const setUser = useAuthStore((s) => s.setUser)
  const navigate = useNavigate()
  const [challenges, setChallenges] = useState<ChallengeResponse[]>([])
  const [heatmap, setHeatmap] = useState<HeatmapEntry[]>([])
  const [loading, setLoading] = useState(true)
  const [checkinTarget, setCheckinTarget] = useState<number | null>(null)
  const [todayStatus, setTodayStatus] = useState<Record<number, boolean>>({})

  useEffect(() => {
    if (!user) return
    setLoading(true)
    Promise.all([
      challengeApi.list(),
      userApi.heatmap(user.id),
    ])
      .then(([ch, hm]) => {
        setChallenges(ch)
        setHeatmap(hm)
        const activeIds = ch.filter((c) => c.status === 'active' || c.status === 'pending').map((c) => c.id)
        return Promise.all(activeIds.map((id) => checkinApi.today(id).then((r) => ({ id, checked: r.checked_in || r.used_freeze }))))
      })
      .then((statuses) => {
        const map: Record<number, boolean> = {}
        statuses.forEach((s) => (map[s.id] = s.checked))
        setTodayStatus(map)
      })
      .catch(() => {})
      .finally(() => setLoading(false))
  }, [user])

  if (!user) return null

  const activeChallenges = challenges.filter((c) => c.status === 'active' || c.status === 'pending')
  const credit = creditLabel(user.credit_score)
  const { current, next } = levelXp(user.level)
  const xpProgress = next > current ? Math.round(((user.xp - current) / (next - current)) * 100) : 100

  return (
    <Spin spinning={loading}>
      <Space direction="vertical" className="w-full" size="large">
        <Row gutter={[16, 16]}>
          <Col xs={24} sm={12} lg={6}>
            <Card>
              <Statistic title="等级" value={user.level} prefix={<ThunderboltOutlined />} suffix={`Lv.${user.level}`} />
              <Progress percent={xpProgress} size="small" showInfo={false} strokeColor="#6366f1" />
              <Typography.Text type="secondary" className="text-xs">{user.xp} / {next} XP</Typography.Text>
            </Card>
          </Col>
          <Col xs={24} sm={12} lg={6}>
            <Card>
              <Statistic title="信用分" value={user.credit_score} suffix="/200" />
              <Tag color={credit.color}>{credit.text}</Tag>
            </Card>
          </Col>
          <Col xs={24} sm={12} lg={6}>
            <Card>
              <Statistic title="进行中的挑战" value={activeChallenges.length} prefix={<FireOutlined />} />
            </Card>
          </Col>
          <Col xs={24} sm={12} lg={6}>
            <Card>
              <Statistic
                title="连击保护"
                value={user.streak_freezes}
                prefix={<SafetyCertificateOutlined />}
                suffix="次"
              />
            </Card>
          </Col>
        </Row>

        <Card
          title="今日待打卡"
          extra={
            <Button type="primary" icon={<PlusOutlined />} onClick={() => navigate('/challenge/create')}>
              创建挑战
            </Button>
          }
        >
          {activeChallenges.length === 0 ? (
            <Empty description="暂无进行中的挑战">
              <Button type="primary" onClick={() => navigate('/challenge/create')}>
                立即创建
              </Button>
            </Empty>
          ) : (
            <Row gutter={[16, 16]}>
              {activeChallenges.map((ch) => (
                <Col xs={24} md={12} xl={8} key={ch.id}>
                  <Card
                    size="small"
                    hoverable
                    className={todayStatus[ch.id] ? 'border-green-300' : 'border-orange-300'}
                    style={{ borderWidth: 2 }}
                  >
                    <div className="flex items-center justify-between mb-2">
                      <Space>
                        <span>{categoryIcon(ch.category)}</span>
                        <Typography.Text strong className="cursor-pointer" onClick={() => navigate(`/challenge/${ch.id}`)}>
                          {ch.title}
                        </Typography.Text>
                      </Space>
                      {todayStatus[ch.id] ? (
                        <Tag icon={<CheckCircleOutlined />} color="success">已打卡</Tag>
                      ) : (
                        <Button
                          type="primary"
                          size="small"
                          onClick={(e) => { e.stopPropagation(); setCheckinTarget(ch.id) }}
                        >
                          打卡
                        </Button>
                      )}
                    </div>
                    <Progress percent={ch.progress} size="small" strokeColor="#6366f1" />
                  </Card>
                </Col>
              ))}
            </Row>
          )}
        </Card>

        <Card title="打卡热力图（近3个月）">
          <HeatmapGrid data={heatmap} />
        </Card>

        {challenges.filter((c) => c.status === 'completed' || c.status === 'failed').length > 0 && (
          <Card title="历史挑战">
            {challenges
              .filter((c) => c.status === 'completed' || c.status === 'failed')
              .map((ch) => (
                <ChallengeCard key={ch.id} challenge={ch} />
              ))}
          </Card>
        )}
      </Space>

      {checkinTarget !== null && (
        <CheckInModal
          challengeId={checkinTarget}
          open
          onClose={() => setCheckinTarget(null)}
          onSuccess={() => {
            setTodayStatus((prev) => ({ ...prev, [checkinTarget!]: true }))
            userApi.stats(user.id).then((s) => setUser(s.user)).catch(() => {})
          }}
        />
      )}
    </Spin>
  )
}

function HeatmapGrid({ data }: { data: HeatmapEntry[] }) {
  const dataMap = new Map(data.map((d) => [d.date, d.count]))
  const today = new Date()
  const cells: { date: string; count: number }[] = []
  for (let i = 89; i >= 0; i--) {
    const d = new Date(today)
    d.setDate(d.getDate() - i)
    const key = d.toISOString().slice(0, 10)
    cells.push({ date: key, count: dataMap.get(key) || 0 })
  }

  return (
    <div className="flex flex-wrap gap-1">
      {cells.map((cell) => (
        <Tooltip key={cell.date} title={`${cell.date}: ${cell.count}次打卡`}>
          <div
            className="rounded-sm"
            style={{
              width: 14,
              height: 14,
              backgroundColor:
                cell.count === 0
                  ? '#ebedf0'
                  : cell.count === 1
                  ? '#9be9a8'
                  : cell.count <= 3
                  ? '#40c463'
                  : '#30a14e',
            }}
          />
        </Tooltip>
      ))}
    </div>
  )
}
