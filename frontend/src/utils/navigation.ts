/**
 * Navigation utility for programmatic navigation outside of React components
 * This is used by stores and other non-component code
 */

let navigateFunction: ((path: string) => void) | null = null;

export const setNavigate = (navigate: (path: string) => void) => {
  navigateFunction = navigate;
};

export const navigateTo = (path: string) => {
  if (navigateFunction) {
    navigateFunction(path);
  } else {
    // Fallback to window.location if navigate is not set
    console.warn('Navigate function not set, falling back to window.location');
    window.location.href = path;
  }
};
