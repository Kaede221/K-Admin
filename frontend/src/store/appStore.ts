import { create } from 'zustand';
import { persist } from 'zustand/middleware';

export interface TabItem {
  key: string;
  label: string;
  path: string;
  closable: boolean;
}

interface AppState {
  collapsed: boolean;
  tabs: TabItem[];
  activeTab: string;

  // Actions
  toggleSidebar: () => void;
  setSidebarCollapsed: (collapsed: boolean) => void;
  addTab: (tab: TabItem) => void;
  removeTab: (key: string) => void;
  setActiveTab: (key: string) => void;
  clearTabs: () => void;
  closeOtherTabs: (key: string) => void;
  refreshTab: (key: string) => void;
}

export const useAppStore = create<AppState>()(
  persist(
    (set, get) => ({
      collapsed: false,
      tabs: [
        {
          key: '/dashboard',
          label: 'Dashboard',
          path: '/dashboard',
          closable: false,
        },
      ],
      activeTab: '/dashboard',

      toggleSidebar: () => {
        set({ collapsed: !get().collapsed });
      },

      setSidebarCollapsed: (collapsed: boolean) => {
        set({ collapsed });
      },

      addTab: (tab: TabItem) => {
        const { tabs } = get();
        const existingTab = tabs.find((t) => t.key === tab.key);

        if (!existingTab) {
          set({
            tabs: [...tabs, tab],
            activeTab: tab.key,
          });
        } else {
          set({ activeTab: tab.key });
        }
      },

      removeTab: (key: string) => {
        const { tabs, activeTab } = get();
        const newTabs = tabs.filter((tab) => tab.key !== key);

        // If removing active tab, switch to the last tab
        if (activeTab === key && newTabs.length > 0) {
          const lastTab = newTabs[newTabs.length - 1];
          set({
            tabs: newTabs,
            activeTab: lastTab.key,
          });
          // Navigate to the last tab
          window.location.href = lastTab.path;
        } else {
          set({ tabs: newTabs });
        }
      },

      setActiveTab: (key: string) => {
        set({ activeTab: key });
      },

      clearTabs: () => {
        // Keep only non-closable tabs (like Dashboard)
        const { tabs } = get();
        const nonClosableTabs = tabs.filter((tab) => !tab.closable);
        set({
          tabs: nonClosableTabs,
          activeTab: nonClosableTabs[0]?.key || '',
        });
      },

      closeOtherTabs: (key: string) => {
        const { tabs } = get();
        const newTabs = tabs.filter((tab) => tab.key === key || !tab.closable);
        set({
          tabs: newTabs,
          activeTab: key,
        });
      },

      refreshTab: (key: string) => {
        // Force remount by removing and re-adding the tab
        const { tabs } = get();
        const tab = tabs.find((t) => t.key === key);
        if (tab) {
          // Trigger a re-render by updating the tab
          const newTabs = tabs.map((t) =>
            t.key === key ? { ...t, key: `${key}-${Date.now()}` } : t
          );
          set({ tabs: newTabs });
          // Restore original key after a brief delay
          setTimeout(() => {
            const restoredTabs = newTabs.map((t) =>
              t.key.startsWith(key) ? { ...t, key } : t
            );
            set({ tabs: restoredTabs });
          }, 100);
        }
      },
    }),
    {
      name: 'app-storage',
      partialize: (state) => ({
        collapsed: state.collapsed,
        tabs: state.tabs,
        activeTab: state.activeTab,
      }),
    }
  )
);
