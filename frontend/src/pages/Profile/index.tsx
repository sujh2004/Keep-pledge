import { useState, useEffect } from 'react'
import { useParams } from 'react-router-dom'
import {
  Card, Avatar, Typography, Progress, Tag, Space, Tabs, Spin, Row, Col,
  Statistic, Tooltip, Upload, message, Button, Input, List,
} from 'antd'
import {
  UserOutlined, FireOutlined, TrophyOutlined, SafetyCertificateOutlined,
  EditOutlined, CameraOutlined, StarOutlined,
} from '@ant-design/icons'
import { useAuthStore } from '../../store/auth'
import { userApi, certificateApi, authApi } from '../../api'
import { creditLabel, levelXp } from '../../utils'
import type { UserStats, HeatmapEntry, AchievementResponse, Certificate } from '../../types'
import ChallengeCard from '../../components/ChallengeCard'
import type { ChallengeResponse } from '../../types'

export default function ProfilePage() {
  const { id } = useParams<{ id: string }>()
  const currentUser = useAuthStore((s) => s.user)
  const setUser = useAuthStore((s) => s.setUser)
  const userId = id ? Number(id) : currentUser?.id
  const isSelf = !id || Number(id) === currentUser?.id

  const [stats, setStats] = useState<UserStats | null>(null)
  const [heatmap, setHeatmap] = useState<HeatmapEntry[]>([])
  const [challenges, setChallenges] = useState<ChallengeResponse[]>([])
  const [achievements, setAchievements] = useState<AchievementResponse[]>([])
  const [certificates, setCertificates] = useState<Certificate[]>([])
  const [loading, setLoading] = useState(true)
  const [editing, setEditing] = useState(false)
  const [newUsername, setNewUsername] = useState('')

  useEffect(() => {
    if (!userId) return
    setLoading(true)
    Promise.all([
      userApi.stats(userId),
      userApi.heatmap(userId),
      userApi.challenges(userId),
      userApi.achievements(userId),
      isSelf ? certificateApi.list() : Promise.resolve([]),
    ])
      .then(([s, h, c, a, certs]) => {
        setStats(s)
        setHeatmap(h)
        setChallenges(c)
        setAchievements(a)
        setCertificates(certs)
      })
      .catch(() => {})
      .finally(() => setLoading(false))
  }, [userId, isSelf])

  if (loading || !stats) return <Spin className="flex justify-center mt-20" />

  const u = stats.user
  const credit = creditLabel(u.credit_score)
  const { current, next } = levelXp(u.level)
  const xpProgress = next > current ? Math.round(((u.xp - current) / (next - current)) * 100) : 100

  const handleAvatarUpload = async (file: File) => {
    try {
      const updated = await userApi.uploadAvatar(file)
      setUser(updated)
      setStats({ ...stats, user: updated })
      message.success('头像已更新')
    } catch {
      message.error('上传失败')
    }
    return false
  }

  const handleUpdateUsername = async () => {
    if (!newUsername.trim()) return
    try {
      const updated = await userApi.updateProfile({ username: newUsername.trim() })
      setUser(updated)
      setStats({ ...stats, user: updated })
      setEditing(false)
      message.success('用户名已更新')
    } catch (err: unknown) {
      message.error(err instanceof Error ? err.message : '更新失败')
    }
  }

  return (
    <Space direction="vertical" className="w-full" size="large">
      <Card>
        <div className="flex items-center gap-6">
          <div className="relative">
            <Avatar size={80} src={u.avatar || undefined} icon={<UserOutlined />} />
            {isSelf && (
              <Upload
                showUploadList={false}
                beforeUpload={handleAvatarUpload}
                accept="image/jpeg,image/png,image/webp"
              >
                <Button
                  size="small"
                  shape="circle"
                  icon={<CameraOutlined />}
                  className="absolute bottom-0 right-0"
                />
              </Upload>
            )}
          </div>
          <div className="flex-1">
            <div className="flex items-center gap-2">
              {editing ? (
                <Space>
                  <Input value={newUsername} onChange={(e) => setNewUsername(e.target.value)} size="small" />
                  <Button size="small" type="primary" onClick={handleUpdateUsername}>保存</Button>
                  <Button size="small" onClick={() => setEditing(false)}>取消</Button>
                </Space>
              ) : (
                <Space>
                  <Typography.Title level={3} style={{ margin: 0 }}>{u.username}</Typography.Title>
                  {isSelf && <EditOutlined className="cursor-pointer text-gray-400" onClick={() => { setEditing(true); setNewUsername(u.username) }} />}
                </Space>
              )}
            </div>
            <Space className="mt-2">
              <Tag color="purple">Lv.{u.level}</Tag>
              <Tag color={credit.color}>{credit.text}</Tag>
              <Typography.Text
                type="secondary"
                copyable={{ text: String(u.id), tooltips: ['复制用户ID', '已复制'] }}
              >
                用户ID：{u.id}
              </Typography.Text>
              <Typography.Text type="secondary">{u.xp} XP</Typography.Text>
            </Space>
            <Progress percent={xpProgress} size="small" className="mt-2 max-w-xs" strokeColor="#6366f1" showInfo={false} />
            <Typography.Text type="secondary" className="text-xs">
              {u.xp} / {next} XP · 距离下一级还需 {next - u.xp} XP
            </Typography.Text>
          </div>
        </div>
      </Card>

      <Row gutter={[16, 16]}>
        <Col xs={12} sm={6}>
          <Card><Statistic title="进行中的挑战" value={stats.active_challenges} prefix={<FireOutlined />} /></Card>
        </Col>
        <Col xs={12} sm={6}>
          <Card><Statistic title="已完成" value={stats.completed} prefix={<TrophyOutlined />} /></Card>
        </Col>
        <Col xs={12} sm={6}>
          <Card><Statistic title="总打卡" value={stats.total_checkins} /></Card>
        </Col>
        <Col xs={12} sm={6}>
          <Card><Statistic title="信用分" value={u.credit_score} suffix="/200" prefix={<SafetyCertificateOutlined />} /></Card>
        </Col>
      </Row>

      <Tabs
        items={[
          {
            key: 'challenges',
            label: '挑战列表',
            children: challenges.length ? challenges.map((c) => <ChallengeCard key={c.id} challenge={c} />) : <Typography.Text type="secondary">暂无挑战</Typography.Text>,
          },
          {
            key: 'achievements',
            label: `成就 (${achievements.filter((a) => a.unlocked).length}/${achievements.length})`,
            children: (
              <Row gutter={[16, 16]}>
                {achievements.map((a) => (
                  <Col xs={12} sm={8} md={6} key={a.id}>
                    <Card
                      size="small"
                      className={a.unlocked ? '' : 'opacity-40'}
                      style={{ textAlign: 'center' }}
                    >
                      <div className="text-3xl mb-2">
                        {a.is_hidden && !a.unlocked ? '❓' : <StarOutlined />}
                      </div>
                      <Typography.Text strong>{a.is_hidden && !a.unlocked ? '???' : a.name}</Typography.Text>
                      <br />
                      <Typography.Text type="secondary" className="text-xs">
                        {a.is_hidden && !a.unlocked ? '隐藏成就' : a.description}
                      </Typography.Text>
                    </Card>
                  </Col>
                ))}
              </Row>
            ),
          },
          {
            key: 'heatmap',
            label: '打卡热力图',
            children: <HeatmapGrid data={heatmap} />,
          },
          ...(isSelf && certificates.length > 0
            ? [{
                key: 'certificates',
                label: `证书 (${certificates.length})`,
                children: (
                  <List
                    dataSource={certificates}
                    renderItem={(cert) => (
                      <List.Item>
                        <List.Item.Meta
                          avatar={<TrophyOutlined style={{ fontSize: 24, color: '#faad14' }} />}
                          title={cert.challenge?.title || '守约证书'}
                          description={`证书编号: ${cert.certificate_no} · 颁发日期: ${cert.issued_at?.slice(0, 10)}`}
                        />
                      </List.Item>
                    )}
                  />
                ),
              }]
            : []),
        ]}
      />
    </Space>
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
              backgroundColor: cell.count === 0 ? '#ebedf0' : cell.count === 1 ? '#9be9a8' : cell.count <= 3 ? '#40c463' : '#30a14e',
            }}
          />
        </Tooltip>
      ))}
    </div>
  )
}
