import { Card, Row, Col, Statistic, Typography } from 'antd';
import { UserOutlined, TeamOutlined, FileTextOutlined, SettingOutlined } from '@ant-design/icons';
import { useUserStore } from '@/store/userStore';

const { Title } = Typography;

export function Dashboard() {
  const userInfo = useUserStore((state) => state.userInfo);

  return (
    <div>
      <Title level={2}>欢迎回来，{userInfo?.nickname || userInfo?.username}！</Title>
      
      <Row gutter={[16, 16]} style={{ marginTop: 24 }}>
        <Col xs={24} sm={12} lg={6}>
          <Card>
            <Statistic
              title="用户总数"
              value={1128}
              prefix={<UserOutlined />}
              valueStyle={{ color: '#3f8600' }}
            />
          </Card>
        </Col>
        
        <Col xs={24} sm={12} lg={6}>
          <Card>
            <Statistic
              title="角色数量"
              value={8}
              prefix={<TeamOutlined />}
              valueStyle={{ color: '#cf1322' }}
            />
          </Card>
        </Col>
        
        <Col xs={24} sm={12} lg={6}>
          <Card>
            <Statistic
              title="菜单数量"
              value={32}
              prefix={<FileTextOutlined />}
              valueStyle={{ color: '#1890ff' }}
            />
          </Card>
        </Col>
        
        <Col xs={24} sm={12} lg={6}>
          <Card>
            <Statistic
              title="系统配置"
              value={15}
              prefix={<SettingOutlined />}
              valueStyle={{ color: '#722ed1' }}
            />
          </Card>
        </Col>
      </Row>

      <Card title="最近活动" style={{ marginTop: 24 }}>
        <p>暂无活动记录</p>
      </Card>
    </div>
  );
}

export default Dashboard;
