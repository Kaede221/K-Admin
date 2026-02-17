import {
  Layout,
  Space,
  Avatar,
  Dropdown,
  Typography,
  Modal,
  message,
} from "antd";
import {
  UserOutlined,
  LogoutOutlined,
  SettingOutlined,
  ExclamationCircleOutlined,
} from "@ant-design/icons";
import type { MenuProps } from "antd";
import { useUserStore } from "@/store/userStore";

const { Header: AntHeader } = Layout;
const { Text } = Typography;

export function Header() {
  const userInfo = useUserStore((state) => state.userInfo);
  const logout = useUserStore((state) => state.logout);

  // User dropdown menu items
  const userMenuItems: MenuProps["items"] = [
    {
      key: "profile",
      icon: <UserOutlined />,
      label: "个人中心",
    },
    {
      key: "settings",
      icon: <SettingOutlined />,
      label: "设置",
    },
    {
      type: "divider",
    },
    {
      key: "logout",
      icon: <LogoutOutlined />,
      label: "退出登录",
      danger: true,
    },
  ];

  // Handle user menu click
  const handleMenuClick: MenuProps["onClick"] = ({ key }) => {
    if (key === "logout") {
      Modal.confirm({
        title: "确认退出",
        icon: <ExclamationCircleOutlined />,
        content: "确定要退出登录吗？",
        okText: "确定",
        cancelText: "取消",
        onOk: () => {
          logout();
          message.success("退出登录成功");
        },
      });
    }
  };

  return (
    <AntHeader
      style={{
        padding: "0 24px",
        background: "#001529",
        display: "flex",
        justifyContent: "flex-end",
        alignItems: "center",
        boxShadow: "0 2px 8px rgba(0,0,0,0.15)",
        height: "56px",
      }}
    >
      <Space size="large">
        <Dropdown
          menu={{ items: userMenuItems, onClick: handleMenuClick }}
          placement="bottomRight"
        >
          <Space style={{ cursor: "pointer" }}>
            <Avatar
              src={userInfo?.headerImg || undefined}
              icon={<UserOutlined />}
              size="default"
              style={{ border: "2px solid rgba(255,255,255,0.2)" }}
            />
            <Text style={{ color: "#fff", fontWeight: 500 }}>
              {userInfo?.nickname || userInfo?.username || "用户"}
            </Text>
          </Space>
        </Dropdown>
      </Space>
    </AntHeader>
  );
}

export default Header;
