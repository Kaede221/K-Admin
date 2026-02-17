export interface MenuItem {
  id: number;
  parentId: number;
  path: string;
  name: string;
  component: string;
  sort: number;
  meta: MenuMeta;
  btn_perms: string[];
  children?: MenuItem[];
}

export interface MenuMeta {
  icon: string;
  title: string;
  hidden: boolean;
  keep_alive: boolean;
}
