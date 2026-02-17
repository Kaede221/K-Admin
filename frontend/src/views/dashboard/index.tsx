import { Card, Row, Col, Statistic, Typography, Spin } from 'antd';
import { UserOutlined, TeamOutlined, FileTextOutlined, SettingOutlined } from '@ant-design/icons';
import { useUserStore } from '@/store/userStore';
import { getDashboardStats, type DashboardStats } from '@/api/dashboard';
import { useEffect, useState } from 'react';
import { message } from 'antd';

const { Title } = Typography;

export function Dashboard() {
  const userInfo = useUserStore((state) => state.userInfo);
  const [stats, setStats] = useState<DashboardStats | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    const fetchStats = async () => {
      try {
        setLoading(true);
        const data = await getDashboardStats();
        setStats(data);
      } catch (error) {
        message.error('获取统计数据失败');
        console.error('Failed to fetch dashboard stats:', error);
      } finally {
        setLoading(false);
      }
    };

    fetchStats();
  }, []);

  return (
    <div>
      <Title level={2}>欢迎回来，{userInfo?.nickname || userInfo?.username}！</Title>
      
      <Spin spinning={loading}>
        <Row gutter={[16, 16]} style={{ marginTop: 24 }}>
          <Col xs={24} sm={12} lg={6}>
            <Card>
              <Statistic
                title="用户总数"
                value={stats?.userCount ?? 0}
                prefix={<UserOutlined />}
                valueStyle={{ color: '#3f8600' }}
              />
            </Card>
          </Col>
          
          <Col xs={24} sm={12} lg={6}>
            <Card>
              <Statistic
                title="角色数量"
                value={stats?.roleCount ?? 0}
                prefix={<TeamOutlined />}
                valueStyle={{ color: '#cf1322' }}
              />
            </Card>
          </Col>
          
          <Col xs={24} sm={12} lg={6}>
            <Card>
              <Statistic
                title="菜单数量"
                value={stats?.menuCount ?? 0}
                prefix={<FileTextOutlined />}
                valueStyle={{ color: '#1890ff' }}
              />
            </Card>
          </Col>
          
          <Col xs={24} sm={12} lg={6}>
            <Card>
              <Statistic
                title="系统配置"
                value={stats?.configCount ?? 0}
                prefix={<SettingOutlined />}
                valueStyle={{ color: '#722ed1' }}
              />
            </Card>
          </Col>
        </Row>
      </Spin>

      <Card title="最近活动" style={{ marginTop: 24 }}>
        <p>暂无活动记录</p>
      </Card>
    </div>
  );
}

export default Dashboard;
