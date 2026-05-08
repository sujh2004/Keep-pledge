import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import {
  Card, Steps, Form, Input, Select, DatePicker, InputNumber,
  Switch, Button, Space, Typography, message, Result,
} from 'antd'
import dayjs from 'dayjs'
import { challengeApi } from '../../api'
import { CATEGORY_LABELS } from '../../types'

const { Step } = Steps
const { TextArea } = Input

export default function CreateChallengePage() {
  const [current, setCurrent] = useState(0)
  const [form] = Form.useForm()
  const [loading, setLoading] = useState(false)
  const [resultId, setResultId] = useState<number | null>(null)
  const navigate = useNavigate()

  const next = async () => {
    try {
      if (current === 0) {
        await form.validateFields(['title', 'description', 'category', 'challenge_type', 'target_days', 'start_date'])
      } else if (current === 1) {
        await form.validateFields(['pledge', 'penalty_type', 'penalty_detail'])
      }
      setCurrent(current + 1)
    } catch {
      // validation failed
    }
  }

  const handleSubmit = async () => {
    try {
      const values = await form.validateFields()
      setLoading(true)
      const res = await challengeApi.create({
        ...values,
        start_date: values.start_date.format('YYYY-MM-DD'),
        is_public: values.is_public || false,
        witness_ids: [],
      })
      setResultId(res.id)
      setCurrent(3)
    } catch (err: unknown) {
      if (err instanceof Error) message.error(err.message)
    } finally {
      setLoading(false)
    }
  }

  if (resultId !== null && current === 3) {
    return (
      <Card>
        <Result
          status="success"
          title="挑战创建成功!"
          subTitle="24小时冷静期内可以撤回挑战"
          extra={[
            <Button type="primary" key="view" onClick={() => navigate(`/challenge/${resultId}`)}>
              查看挑战
            </Button>,
            <Button key="home" onClick={() => navigate('/')}>
              返回首页
            </Button>,
          ]}
        />
      </Card>
    )
  }

  return (
    <Card title="创建挑战">
      <Steps current={current} className="mb-8">
        <Step title="基本信息" />
        <Step title="誓约设定" />
        <Step title="确认发布" />
      </Steps>

      <Form form={form} layout="vertical" className="max-w-2xl mx-auto">
        <div style={{ display: current === 0 ? 'block' : 'none' }}>
          <Form.Item name="title" label="挑战标题" rules={[{ required: true, message: '请输入标题' }, { min: 2, max: 128 }]}>
            <Input placeholder="例如：每天跑步5公里" />
          </Form.Item>
          <Form.Item name="description" label="挑战描述">
            <TextArea rows={3} placeholder="描述你的挑战目标..." maxLength={2000} showCount />
          </Form.Item>
          <Form.Item name="category" label="分类" rules={[{ required: true, message: '请选择分类' }]}>
            <Select placeholder="选择分类">
              {Object.entries(CATEGORY_LABELS).map(([k, v]) => (
                <Select.Option key={k} value={k}>{v}</Select.Option>
              ))}
            </Select>
          </Form.Item>
          <Form.Item name="challenge_type" label="挑战类型" rules={[{ required: true }]} initialValue="daily">
            <Select>
              <Select.Option value="daily">每日打卡</Select.Option>
              <Select.Option value="total">累计打卡</Select.Option>
            </Select>
          </Form.Item>
          <Form.Item name="target_days" label="目标天数" rules={[{ required: true, message: '请输入天数' }]} initialValue={30}>
            <InputNumber min={1} max={365} className="w-full" />
          </Form.Item>
          <Form.Item name="start_date" label="开始日期" rules={[{ required: true, message: '请选择日期' }]}>
            <DatePicker className="w-full" disabledDate={(d) => d.isBefore(dayjs(), 'day')} />
          </Form.Item>
        </div>

        <div style={{ display: current === 1 ? 'block' : 'none' }}>
          <Form.Item name="pledge" label="誓约内容" rules={[{ required: true, message: '请填写誓约' }, { min: 5 }]}>
            <TextArea rows={3} placeholder="例如：如果我未完成挑战，我将向慈善机构捐款100元" />
          </Form.Item>
          <Form.Item name="penalty_type" label="失败后果类型" rules={[{ required: true }]}>
            <Select placeholder="选择后果类型">
              <Select.Option value="donate">公益捐款</Select.Option>
              <Select.Option value="social">社交惩罚</Select.Option>
              <Select.Option value="custom">自定义</Select.Option>
            </Select>
          </Form.Item>
          <Form.Item name="penalty_detail" label="后果详情">
            <Input placeholder="例如：向红十字会捐款100元" />
          </Form.Item>
          <Form.Item name="is_public" label="公开挑战" valuePropName="checked">
            <Switch checkedChildren="公开" unCheckedChildren="私密" />
          </Form.Item>
        </div>

        <div style={{ display: current === 2 ? 'block' : 'none' }}>
          <Typography.Title level={4}>确认信息</Typography.Title>
          <Typography.Paragraph type="secondary">
            请确认以下信息无误后发布挑战。创建后有 24 小时冷静期，期间可以撤回。
          </Typography.Paragraph>
          <div className="bg-gray-50 p-4 rounded-lg mb-4">
            <Space direction="vertical">
              <Typography.Text>标题：{form.getFieldValue('title')}</Typography.Text>
              <Typography.Text>目标：{form.getFieldValue('target_days')} 天</Typography.Text>
              <Typography.Text>誓约：{form.getFieldValue('pledge')}</Typography.Text>
            </Space>
          </div>
        </div>

        <div className="flex justify-between mt-6">
          {current > 0 && (
            <Button onClick={() => setCurrent(current - 1)}>上一步</Button>
          )}
          <div className="ml-auto">
            {current < 2 ? (
              <Button type="primary" onClick={next}>下一步</Button>
            ) : (
              <Button type="primary" loading={loading} onClick={handleSubmit}>
                发布挑战
              </Button>
            )}
          </div>
        </div>
      </Form>
    </Card>
  )
}
