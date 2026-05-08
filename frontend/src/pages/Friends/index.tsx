import { useState, useEffect } from 'react'
import {
  Card, List, Avatar, Tag, Button, Space, Typography, Input,
  InputNumber, Spin, message, Empty, Popconfirm,
} from 'antd'
import { UserOutlined, UserAddOutlined, UserDeleteOutlined } from '@ant-design/icons'
import { socialApi } from '../../api'
import type { Friendship } from '../../types'

export default function FriendsPage() {
  const [friends, setFriends] = useState<Friendship[]>([])
  const [loading, setLoading] = useState(true)
  const [friendId, setFriendId] = useState<number | null>(null)
  const [sending, setSending] = useState(false)

  const load = () => {
    setLoading(true)
    socialApi.friends()
      .then(setFriends)
      .catch(() => {})
      .finally(() => setLoading(false))
  }

  useEffect(() => { load() }, [])

  const handleRequest = async () => {
    if (!friendId) return
    setSending(true)
    try {
      await socialApi.request(friendId)
      message.success('好友请求已发送')
      setFriendId(null)
      load()
    } catch (err: unknown) {
      message.error(err instanceof Error ? err.message : '发送失败')
    } finally {
      setSending(false)
    }
  }

  const handleAccept = async (id: number) => {
    try {
      await socialApi.accept(id)
      message.success('已接受好友请求')
      load()
    } catch (err: unknown) {
      message.error(err instanceof Error ? err.message : '操作失败')
    }
  }

  const handleDelete = async (id: number) => {
    try {
      await socialApi.delete(id)
      message.success('好友已删除')
      load()
    } catch (err: unknown) {
      message.error(err instanceof Error ? err.message : '操作失败')
    }
  }

  const accepted = friends.filter((f) => f.status === 'accepted')
  const incomingPending = friends.filter((f) => f.status === 'pending' && f.direction === 'incoming')
  const outgoingPending = friends.filter((f) => f.status === 'pending' && f.direction === 'outgoing')

  return (
    <Space direction="vertical" className="w-full" size="large">
      <Card title="添加好友">
        <Space>
          <InputNumber
            placeholder="输入用户 ID"
            value={friendId}
            onChange={(v) => setFriendId(v)}
            min={1}
            className="w-40"
          />
          <Button
            type="primary"
            icon={<UserAddOutlined />}
            onClick={handleRequest}
            loading={sending}
            disabled={!friendId}
          >
            发送请求
          </Button>
        </Space>
      </Card>

      {incomingPending.length > 0 && (
        <Card title={`收到的好友请求 (${incomingPending.length})`}>
          <List
            dataSource={incomingPending}
            renderItem={(item) => (
              <List.Item
                actions={[
                  <Button type="primary" size="small" onClick={() => handleAccept(item.id)} key="accept">
                    接受
                  </Button>,
                ]}
              >
                <List.Item.Meta
                  avatar={<Avatar src={item.friend.avatar || undefined} icon={<UserOutlined />} />}
                  title={item.friend.username}
                  description={<Tag color="orange">等待你接受</Tag>}
                />
              </List.Item>
            )}
          />
        </Card>
      )}

      {outgoingPending.length > 0 && (
        <Card title={`已发送的好友请求 (${outgoingPending.length})`}>
          <List
            dataSource={outgoingPending}
            renderItem={(item) => (
              <List.Item
                actions={[
                  <Popconfirm title="确定取消这条好友请求?" onConfirm={() => handleDelete(item.id)} key="cancel">
                    <Button size="small">取消请求</Button>
                  </Popconfirm>,
                ]}
              >
                <List.Item.Meta
                  avatar={<Avatar src={item.friend.avatar || undefined} icon={<UserOutlined />} />}
                  title={item.friend.username}
                  description={<Tag color="default">等待对方接受</Tag>}
                />
              </List.Item>
            )}
          />
        </Card>
      )}

      <Card title={`我的好友 (${accepted.length})`}>
        <Spin spinning={loading}>
          {accepted.length === 0 ? (
            <Empty description="还没有好友，快去添加吧" />
          ) : (
            <List
              dataSource={accepted}
              renderItem={(item) => (
                <List.Item
                  actions={[
                    <Popconfirm title="确定删除好友?" onConfirm={() => handleDelete(item.id)} key="del">
                      <Button danger size="small" icon={<UserDeleteOutlined />}>
                        删除
                      </Button>
                    </Popconfirm>,
                  ]}
                >
                  <List.Item.Meta
                    avatar={<Avatar src={item.friend.avatar || undefined} icon={<UserOutlined />} />}
                    title={item.friend.username}
                    description={
                      <Space>
                        <Tag color="blue">Lv.{item.friend.level}</Tag>
                        <Typography.Text type="secondary">{item.friend.xp} XP</Typography.Text>
                      </Space>
                    }
                  />
                </List.Item>
              )}
            />
          )}
        </Spin>
      </Card>
    </Space>
  )
}
