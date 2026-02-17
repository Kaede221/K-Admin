import { Switch } from 'antd';
import { BulbOutlined, BulbFilled } from '@ant-design/icons';
import { useAppStore } from '@/store/appStore';

/**
 * ThemeSwitch component - Toggle between light and dark themes
 * 
 * Integrates with Ant Design ConfigProvider to update theme tokens
 * Persists theme preference to localStorage
 */
export function ThemeSwitch() {
  const theme = useAppStore((state) => state.theme);
  const toggleTheme = useAppStore((state) => state.toggleTheme);

  const isDark = theme === 'dark';

  return (
    <Switch
      checked={isDark}
      onChange={toggleTheme}
      checkedChildren={<BulbFilled />}
      unCheckedChildren={<BulbOutlined />}
      title={isDark ? '切换到亮色模式' : '切换到暗色模式'}
    />
  );
}

export default ThemeSwitch;
