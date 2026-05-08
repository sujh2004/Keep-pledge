import { useState, useEffect } from 'react'
import { useParams } from 'react-router-dom'
import { Card, Typography, Descriptions, Spin, Space, Tag, Divider, Image } from 'antd'
import { SafetyCertificateOutlined, TrophyOutlined } from '@ant-design/icons'
import { certificateApi } from '../../api'
import type { Certificate } from '../../types'

export default function CertificatePage() {
  const { id } = useParams<{ id: string }>()
  const [cert, setCert] = useState<Certificate | null>(null)
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    if (!id) return
    certificateApi.get(Number(id))
      .then(setCert)
      .catch(() => {})
      .finally(() => setLoading(false))
  }, [id])

  if (loading) return <Spin className="flex justify-center mt-20" />
  if (!cert) return <Typography.Text type="secondary">证书不存在</Typography.Text>

  return (
    <Card className="max-w-2xl mx-auto">
      <div className="text-center mb-6">
        <TrophyOutlined style={{ fontSize: 48, color: '#faad14' }} />
        <Typography.Title level={2} style={{ color: '#6366f1' }}>
          守约证书
        </Typography.Title>
        <Divider />
      </div>
      <div className="bg-gradient-to-br from-indigo-50 to-purple-50 p-8 rounded-lg">
        <div className="text-center mb-6">
          <SafetyCertificateOutlined style={{ fontSize: 64, color: '#6366f1' }} />
          <Typography.Title level={3} className="mt-4">
            {cert.challenge?.title}
          </Typography.Title>
          <Tag color="gold" className="text-lg px-4 py-1">挑战完成</Tag>
        </div>
        <Descriptions bordered column={1} size="small" className="bg-white rounded-lg">
          <Descriptions.Item label="证书编号">{cert.certificate_no}</Descriptions.Item>
          <Descriptions.Item label="挑战名称">{cert.challenge?.title}</Descriptions.Item>
          <Descriptions.Item label="颁发日期">{cert.issued_at?.slice(0, 10)}</Descriptions.Item>
          <Descriptions.Item label="挑战描述">{cert.challenge?.description}</Descriptions.Item>
        </Descriptions>
        {cert.image_url && (
          <div className="mt-6 text-center">
            <Image
              src={cert.image_url}
              alt="守约证书图片"
              className="rounded-lg border border-indigo-100"
              style={{ maxHeight: 360 }}
            />
          </div>
        )}
      </div>
    </Card>
  )
}
