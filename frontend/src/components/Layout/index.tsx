import { useState, useEffect } from 'react'
import { useNavigate, useLocation } from 'react-router-dom'
import { Layout, Menu, Avatar, Badge, Dropdown, Space, Typography } from 'antd'
import {
  HomeOutlined,
  TrophyOutlined,
  PlusCircleOutlined,
  TeamOutlined,
  CrownOutlined,
  CompassOutlined,
  UserOutlined,
  StarOutlined,
  LogoutOutlined,
  BellOutlined,
} from '@ant-design/icons'
import { useAuthStore } from '../../store/auth'
import { notificationApi } from '../../api'

const { Header, Content, Sider } = Layout

const menuItems = [
  { key: '/', icon: <HomeOutlined />, label: '首页' },
  { key: '/explore', icon: <CompassOutlined />, label: '挑战广场' },
  { key: '/challenge/create', icon: <PlusCircleOutlined />, label: '创建挑战' },
  { key: '/leaderboard', icon: <CrownOutlined />, label: '排行榜' },
  { key: '/achievements', icon: <StarOutlined />, label: '成就墙' },
  { key: '/friends', icon: <TeamOutlined />, label: '好友' },
  { key: '/notifications', icon: <BellOutlined />, label: '通知' },
  { key: '/profile', icon: <UserOutlined />, label: '个人中心' },
]

export default function AppLayout({ children }: { children: React.ReactNode }) {
  const navigate = useNavigate()
  const location = useLocation()
  const user = useAuthStore((s) => s.user)
  const logout = useAuthStore((s) => s.logout)
  const [collapsed, setCollapsed] = useState(false)
  const [unread, setUnread] = useState(0)

  useEffect(() => {
    const poll = () => {
      notificationApi.unreadCount().then((r) => setUnread(r.count)).catch(() => {})
    }
    poll()
    const timer = setInterval(poll, 30000)
    return () => clearInterval(timer)
  }, [])

  const selectedKey = menuItems.find((item) => {
    if (item.key === '/') return location.pathname === '/'
    return location.pathname.startsWith(item.key)
  })?.key || '/'

  return (
    <Layout style={{ minHeight: '100vh' }}>
      <Sider
        collapsible
        collapsed={collapsed}
        onCollapse={setCollapsed}
        breakpoint="lg"
        theme="light"
        style={{ borderRight: '1px solid #f0f0f0' }}
      >
        <div className="flex items-center justify-center py-4 px-2">
          <TrophyOutlined style={{ fontSize: 24, color: '#6366f1' }} />
          {!collapsed && (
            <Typography.Title level={4} style={{ margin: '0 0 0 8px', color: '#6366f1' }}>
              守约挑战
            </Typography.Title>
          )}
        </div>
        <Menu
          mode="inline"
          selectedKeys={[selectedKey]}
          items={menuItems.map((item) => ({
            ...item,
            label:
              item.key === '/notifications' ? (
                <Badge count={unread} size="small" offset={[8, 0]}>
                  {item.label}
                </Badge>
              ) : (
                item.label
              ),
          }))}
          onClick={({ key }) => navigate(key)}
        />
      </Sider>
      <Layout>
        <Header className="flex items-center justify-end bg-white px-6" style={{ borderBottom: '1px solid #f0f0f0' }}>
          <Space size="middle">
            <Badge count={unread} size="small">
              <BellOutlined
                style={{ fontSize: 20, cursor: 'pointer' }}
                onClick={() => navigate('/notifications')}
              />
            </Badge>
            <Dropdown
              menu={{
                items: [
                  { key: 'profile', icon: <UserOutlined />, label: '个人中心', onClick: () => navigate('/profile') },
                  { type: 'divider' },
                  { key: 'logout', icon: <LogoutOutlined />, label: '退出登录', danger: true, onClick: () => { logout(); navigate('/login') } },
                ],
              }}
            >
              <Space style={{ cursor: 'pointer' }}>
                <Avatar src={user?.avatar || undefined} icon={<UserOutlined />} />
                <span>{user?.username}</span>
              </Space>
            </Dropdown>
          </Space>
        </Header>
        <Content className="p-6">{children}</Content>
      </Layout>
    </Layout>
  )
}
