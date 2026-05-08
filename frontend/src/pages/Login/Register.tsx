import { useState } from 'react'
import { useNavigate, Link } from 'react-router-dom'
import { Form, Input, Button, Card, Typography, message, Space } from 'antd'
import { UserOutlined, MailOutlined, LockOutlined, TrophyOutlined } from '@ant-design/icons'
import { authApi } from '../../api'
import { useAuthStore } from '../../store/auth'

export default function RegisterPage() {
  const [loading, setLoading] = useState(false)
  const navigate = useNavigate()
  const setAuth = useAuthStore((s) => s.setAuth)

  const onFinish = async (values: { username: string; email: string; password: string }) => {
    setLoading(true)
    try {
      const res = await authApi.register(values)
      setAuth(res.token, res.refresh_token, res.user)
      message.success('注册成功')
      navigate('/')
    } catch (err: unknown) {
      message.error(err instanceof Error ? err.message : '注册失败')
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="min-h-screen flex items-center justify-center bg-gradient-to-br from-indigo-50 to-purple-50">
      <Card className="w-full max-w-md shadow-lg" bordered={false}>
        <div className="text-center mb-8">
          <TrophyOutlined style={{ fontSize: 48, color: '#6366f1' }} />
          <Typography.Title level={2} style={{ marginTop: 16, color: '#6366f1' }}>
            注册账号
          </Typography.Title>
        </div>
        <Form layout="vertical" onFinish={onFinish} size="large">
          <Form.Item name="username" rules={[{ required: true, message: '请输入用户名' }, { min: 3, max: 32, message: '3-32个字符' }]}>
            <Input prefix={<UserOutlined />} placeholder="用户名" />
          </Form.Item>
          <Form.Item name="email" rules={[{ required: true, message: '请输入邮箱' }, { type: 'email', message: '邮箱格式不正确' }]}>
            <Input prefix={<MailOutlined />} placeholder="邮箱" />
          </Form.Item>
          <Form.Item name="password" rules={[{ required: true, message: '请输入密码' }, { min: 8, message: '密码至少8位' }]}>
            <Input.Password prefix={<LockOutlined />} placeholder="密码" />
          </Form.Item>
          <Form.Item
            name="confirm"
            dependencies={['password']}
            rules={[
              { required: true, message: '请确认密码' },
              ({ getFieldValue }) => ({
                validator(_, value) {
                  if (!value || getFieldValue('password') === value) return Promise.resolve()
                  return Promise.reject(new Error('两次密码输入不一致'))
                },
              }),
            ]}
          >
            <Input.Password prefix={<LockOutlined />} placeholder="确认密码" />
          </Form.Item>
          <Form.Item>
            <Button type="primary" htmlType="submit" loading={loading} block>
              注册
            </Button>
          </Form.Item>
        </Form>
        <div className="text-center">
          <Space>
            <Typography.Text type="secondary">已有账号?</Typography.Text>
            <Link to="/login">去登录</Link>
          </Space>
        </div>
      </Card>
    </div>
  )
}
