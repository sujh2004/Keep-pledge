import { useState, useEffect } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import {
  Card, Typography, Progress, Tag, Space, Button, Avatar, List, Tooltip,
  Spin, Popconfirm, message, Descriptions, Divider, Input, Modal,
} from 'antd'
import {
  FireOutlined, ClockCircleOutlined, UserOutlined, LikeOutlined,
  LikeFilled, CommentOutlined, WarningOutlined, CheckCircleOutlined,
} from '@ant-design/icons'
import dayjs from 'dayjs'
import { challengeApi, checkinApi } from '../../api'
import { useAuthStore } from '../../store/auth'
import { statusLabel, categoryIcon } from '../../utils'
import { CATEGORY_LABELS } from '../../types'
import type { ChallengeDetail, CheckInResponse, Interaction } from '../../types'
import CheckInModal from '../../components/CheckInModal'

export default function ChallengePage() {
  const { id } = useParams<{ id: string }>()
  const user = useAuthStore((s) => s.user)
  const navigate = useNavigate()
  const [detail, setDetail] = useState<ChallengeDetail | null>(null)
  const [loading, setLoading] = useState(true)
  const [checkinOpen, setCheckinOpen] = useState(false)
  const [todayDone, setTodayDone] = useState(false)
  const [commentTarget, setCommentTarget] = useState<number | null>(null)
  const [commentText, setCommentText] = useState('')
  const [interactions, setInteractions] = useState<Record<number, Interaction[]>>({})

  const challengeId = Number(id)

  const load = () => {
    setLoading(true)
    challengeApi
      .get(challengeId)
      .then((d) => {
        setDetail(d)
        return checkinApi.today(challengeId)
      })
      .then((t) => setTodayDone(t.checked_in || t.used_freeze))
      .catch((err) => message.error(err.message))
      .finally(() => setLoading(false))
  }

  useEffect(() => {
    if (challengeId) load()
  }, [challengeId])

  if (loading || !detail) return <Spin className="flex justify-center mt-20" />

  const ch = detail.challenge
  const status = statusLabel(ch.status)
  const isCreator = user?.id === ch.creator_id
  const canCancel = isCreator && dayjs().isBefore(dayjs(ch.can_cancel_before))
  const isActive = ch.status === 'active' || ch.status === 'pending'

  const handleCancel = async () => {
    try {
      await challengeApi.cancel(challengeId)
      message.success('挑战已撤回')
      load()
    } catch (err: unknown) {
      message.error(err instanceof Error ? err.message : '撤回失败')
    }
  }

  const handleLike = async (checkinId: number, liked: boolean) => {
    try {
      if (liked) {
        await checkinApi.unlike(checkinId)
      } else {
        await checkinApi.like(checkinId)
      }
      load()
    } catch {
      // ignore
    }
  }

  const handleComment = async () => {
    if (!commentTarget || !commentText.trim()) return
    try {
      await checkinApi.comment(commentTarget, commentText)
      setCommentText('')
      setCommentTarget(null)
      message.success('评论成功')
      const ints = await checkinApi.interactions(commentTarget)
      setInteractions((prev) => ({ ...prev, [commentTarget]: ints }))
    } catch (err: unknown) {
      message.error(err instanceof Error ? err.message : '评论失败')
    }
  }

  const loadInteractions = async (checkinId: number) => {
    const ints = await checkinApi.interactions(checkinId)
    setInteractions((prev) => ({ ...prev, [checkinId]: ints }))
  }

  return (
    <Space direction="vertical" className="w-full" size="large">
      <Card>
        <div className="flex items-center justify-between mb-4">
          <Space>
            <span className="text-2xl">{categoryIcon(ch.category)}</span>
            <Typography.Title level={3} style={{ margin: 0 }}>{ch.title}</Typography.Title>
            <Tag color={status.color}>{status.text}</Tag>
          </Space>
          <Space>
            {isActive && !todayDone && (
              <Button type="primary" icon={<CheckCircleOutlined />} onClick={() => setCheckinOpen(true)}>
                打卡
              </Button>
            )}
            {todayDone && <Tag color="success" icon={<CheckCircleOutlined />}>今日已打卡</Tag>}
            {canCancel && (
              <Popconfirm title="确定撤回?" onConfirm={handleCancel}>
                <Button danger>撤回挑战</Button>
              </Popconfirm>
            )}
          </Space>
        </div>

        {canCancel && (
          <div className="bg-blue-50 p-3 rounded mb-4 flex items-center gap-2">
            <ClockCircleOutlined style={{ color: '#1890ff' }} />
            <Typography.Text type="secondary">
              冷静期至 {dayjs(ch.can_cancel_before).format('YYYY-MM-DD HH:mm')}，期间可撤回挑战
            </Typography.Text>
          </div>
        )}

        <div className="bg-orange-50 p-4 rounded-lg mb-4 border border-orange-200">
          <Typography.Text strong>誓约：</Typography.Text>
          <Typography.Text>{ch.pledge}</Typography.Text>
          {ch.penalty_detail && (
            <div className="mt-1">
              <Typography.Text type="secondary">
                <WarningOutlined /> 失败后果：{ch.penalty_detail}
              </Typography.Text>
            </div>
          )}
        </div>

        <Descriptions column={{ xs: 1, sm: 2, lg: 3 }} bordered size="small">
          <Descriptions.Item label="分类">{CATEGORY_LABELS[ch.category as keyof typeof CATEGORY_LABELS]}</Descriptions.Item>
          <Descriptions.Item label="类型">{ch.challenge_type === 'daily' ? '每日打卡' : '累计打卡'}</Descriptions.Item>
          <Descriptions.Item label="目标天数">{ch.target_days}天</Descriptions.Item>
          <Descriptions.Item label="开始日期">{ch.start_date}</Descriptions.Item>
          <Descriptions.Item label="结束日期">{ch.end_date}</Descriptions.Item>
          <Descriptions.Item label="完成奖励">{ch.xp_reward} XP</Descriptions.Item>
        </Descriptions>

        <div className="mt-4">
          <Typography.Text type="secondary">完成进度</Typography.Text>
          <Progress percent={ch.progress} strokeColor="#6366f1" />
        </div>
      </Card>

      <Card title={`参与者 (${detail.participants.length})`}>
        <div className="flex flex-wrap gap-4">
          {detail.participants.map((p) => (
            <Card key={p.id} size="small" className="w-48">
              <Space>
                <Avatar src={p.user.avatar || undefined} icon={<UserOutlined />} />
                <div>
                  <Typography.Text strong>{p.user.username}</Typography.Text>
                  <br />
                  <Tag color={p.role === 'creator' ? 'purple' : p.role === 'witness' ? 'cyan' : 'blue'}>
                    {p.role === 'creator' ? '创建者' : p.role === 'witness' ? '见证人' : '参与者'}
                  </Tag>
                  <br />
                  <Typography.Text type="secondary" className="text-xs">
                    <FireOutlined /> {p.current_streak}天连击
                  </Typography.Text>
                </div>
              </Space>
            </Card>
          ))}
        </div>
      </Card>

      <Card title="打卡记录">
        {detail.checkins.length === 0 ? (
          <Typography.Text type="secondary">暂无打卡记录</Typography.Text>
        ) : (
          <List
            dataSource={detail.checkins}
            renderItem={(ci: CheckInResponse) => (
              <List.Item
                actions={[
                  <Tooltip title="点赞" key="like">
                    <Space
                      className="cursor-pointer"
                      onClick={() => handleLike(ci.id, false)}
                    >
                      <LikeOutlined />
                      {ci.like_count}
                    </Space>
                  </Tooltip>,
                  <Tooltip title="评论" key="comment">
                    <Space
                      className="cursor-pointer"
                      onClick={() => {
                        setCommentTarget(ci.id)
                        loadInteractions(ci.id)
                      }}
                    >
                      <CommentOutlined />
                    </Space>
                  </Tooltip>,
                ]}
              >
                <List.Item.Meta
                  avatar={<Avatar src={ci.user.avatar || undefined} icon={<UserOutlined />} />}
                  title={
                    <Space>
                      <span>{ci.user.username}</span>
                      <Tag><FireOutlined /> {ci.streak_count}天</Tag>
                      <Typography.Text type="secondary" className="text-xs">
                        +{ci.xp_earned} XP
                      </Typography.Text>
                    </Space>
                  }
                  description={
                    <div>
                      <Typography.Paragraph>{ci.content}</Typography.Paragraph>
                      {ci.image_url && (
                        <img
                          src={ci.image_url}
                          alt="打卡图片"
                          className="max-w-xs rounded-lg"
                          style={{ maxHeight: 200 }}
                        />
                      )}
                      <div className="text-xs text-gray-400 mt-1">{ci.checkin_date}</div>
                    </div>
                  }
                />
              </List.Item>
            )}
          />
        )}
      </Card>

      <CheckInModal
        challengeId={challengeId}
        open={checkinOpen}
        onClose={() => setCheckinOpen(false)}
        onSuccess={() => {
          setTodayDone(true)
          load()
        }}
      />

      <Modal
        title="评论"
        open={commentTarget !== null}
        onCancel={() => { setCommentTarget(null); setCommentText('') }}
        onOk={handleComment}
        okText="发送"
      >
        {commentTarget && interactions[commentTarget] && (
          <List
            size="small"
            dataSource={interactions[commentTarget].filter((i) => i.type === 'comment')}
            renderItem={(item) => (
              <List.Item>
                <List.Item.Meta
                  avatar={<Avatar src={item.user.avatar || undefined} icon={<UserOutlined />} size="small" />}
                  title={item.user.username}
                  description={item.content}
                />
              </List.Item>
            )}
            className="mb-4 max-h-60 overflow-auto"
          />
        )}
        <Input.TextArea
          rows={3}
          value={commentText}
          onChange={(e) => setCommentText(e.target.value)}
          placeholder="写下你的评论..."
          maxLength={512}
        />
      </Modal>
    </Space>
  )
}
