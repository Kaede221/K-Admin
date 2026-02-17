import { useState } from 'react';
import { Form, Input, Button, Card, message } from 'antd';
import { UserOutlined, LockOutlined } from '@ant-design/icons';
import { useNavigate } from 'react-router-dom';
import { useUserStore } from '@/store/userStore';

interface LoginForm {
  username: string;
  password: string;
}

export function Login() {
  const navigate = useNavigate();
  const [loading, setLoading] = useState(false);
  const login = useUserStore((state) => state.login);

  const handleSubmit = async (values: LoginForm) => {
    setLoading(true);
    try {
      await login(values.username, values.password);
      message.success('登录成功');
      navigate('/dashboard');
    } catch (error: any) {
      message.error(error.message || '登录失败');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div
      style={{
        display: 'flex',
        justifyContent: 'center',
        alignItems: 'center',
        minHeight: '100vh',
        background: 'linear-gradient(135deg, #ECFEFF 0%, #E0F2FE 50%, #F0FDFA 100%)',
        position: 'relative',
        overflow: 'hidden',
      }}
    >
      {/* Decorative elements */}
      <div
        style={{
          position: 'absolute',
          top: '-10%',
          right: '-5%',
          width: '500px',
          height: '500px',
          background: 'radial-gradient(circle, rgba(8, 145, 178, 0.1) 0%, transparent 70%)',
          borderRadius: '50%',
          pointerEvents: 'none',
        }}
      />
      <div
        style={{
          position: 'absolute',
          bottom: '-10%',
          left: '-5%',
          width: '400px',
          height: '400px',
          background: 'radial-gradient(circle, rgba(34, 211, 238, 0.08) 0%, transparent 70%)',
          borderRadius: '50%',
          pointerEvents: 'none',
        }}
      />

      <Card
        style={{
          width: 440,
          maxWidth: '90vw',
          background: 'rgba(255, 255, 255, 0.95)',
          backdropFilter: 'blur(10px)',
          border: '1px solid rgba(8, 145, 178, 0.1)',
          borderRadius: '16px',
          boxShadow: '0 8px 32px rgba(8, 145, 178, 0.12)',
        }}
        bodyStyle={{ padding: '48px 40px' }}
        bordered={false}
      >
        {/* Title */}
        <div style={{ textAlign: 'center', marginBottom: '40px' }}>
          <h1
            style={{
              fontSize: '28px',
              fontWeight: 600,
              color: '#164E63',
              margin: '0 0 8px 0',
              fontFamily: "'Fira Sans', sans-serif",
            }}
          >
            K-Admin 管理系统
          </h1>
          <p
            style={{
              fontSize: '14px',
              color: '#0891B2',
              margin: 0,
              fontFamily: "'Fira Sans', sans-serif",
            }}
          >
            欢迎回来，请登录您的账户
          </p>
        </div>

        <Form
          name="login"
          onFinish={handleSubmit}
          autoComplete="off"
          size="large"
          layout="vertical"
        >
          <Form.Item
            label={<span style={{ color: '#164E63', fontWeight: 500 }}>用户名</span>}
            name="username"
            rules={[{ required: true, message: '请输入用户名' }]}
          >
            <Input
              prefix={<UserOutlined style={{ color: '#0891B2' }} />}
              placeholder="请输入用户名"
              style={{
                borderRadius: '8px',
                border: '1px solid #E0F2FE',
                transition: 'all 0.2s',
              }}
              onFocus={(e) => {
                e.target.style.borderColor = '#0891B2';
                e.target.style.boxShadow = '0 0 0 2px rgba(8, 145, 178, 0.1)';
              }}
              onBlur={(e) => {
                e.target.style.borderColor = '#E0F2FE';
                e.target.style.boxShadow = 'none';
              }}
            />
          </Form.Item>

          <Form.Item
            label={<span style={{ color: '#164E63', fontWeight: 500 }}>密码</span>}
            name="password"
            rules={[{ required: true, message: '请输入密码' }]}
            style={{ marginBottom: '32px' }}
          >
            <Input.Password
              prefix={<LockOutlined style={{ color: '#0891B2' }} />}
              placeholder="请输入密码"
              style={{
                borderRadius: '8px',
                border: '1px solid #E0F2FE',
                transition: 'all 0.2s',
              }}
              onFocus={(e) => {
                e.target.style.borderColor = '#0891B2';
                e.target.style.boxShadow = '0 0 0 2px rgba(8, 145, 178, 0.1)';
              }}
              onBlur={(e) => {
                e.target.style.borderColor = '#E0F2FE';
                e.target.style.boxShadow = 'none';
              }}
            />
          </Form.Item>

          <Form.Item style={{ marginBottom: 0 }}>
            <Button
              type="primary"
              htmlType="submit"
              loading={loading}
              block
              style={{
                height: '48px',
                fontSize: '16px',
                fontWeight: 500,
                borderRadius: '8px',
                background: 'linear-gradient(135deg, #22C55E 0%, #16A34A 100%)',
                border: 'none',
                boxShadow: '0 4px 12px rgba(34, 197, 94, 0.3)',
                transition: 'all 0.2s',
                cursor: 'pointer',
              }}
              onMouseEnter={(e) => {
                e.currentTarget.style.transform = 'translateY(-2px)';
                e.currentTarget.style.boxShadow = '0 6px 16px rgba(34, 197, 94, 0.4)';
              }}
              onMouseLeave={(e) => {
                e.currentTarget.style.transform = 'translateY(0)';
                e.currentTarget.style.boxShadow = '0 4px 12px rgba(34, 197, 94, 0.3)';
              }}
            >
              登录
            </Button>
          </Form.Item>
        </Form>

        {/* Footer */}
        <div
          style={{
            marginTop: '32px',
            paddingTop: '24px',
            borderTop: '1px solid #E0F2FE',
            textAlign: 'center',
          }}
        >
          <p
            style={{
              fontSize: '13px',
              color: '#0891B2',
              margin: 0,
              fontFamily: "'Fira Sans', sans-serif",
            }}
          >
            © 2026 K-Admin. 保留所有权利
          </p>
        </div>
      </Card>
    </div>
  );
}

export default Login;
