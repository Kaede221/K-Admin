import { describe, it, expect, beforeEach, vi } from 'vitest';
import { render, screen } from '@testing-library/react';
import { AuthButton } from './AuthButton';
import { useUserStore } from '@/store/userStore';

// Mock the userStore
vi.mock('@/store/userStore', () => ({
  useUserStore: vi.fn(),
}));

describe('AuthButton', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('should render button when user has permission', () => {
    // Mock hasPermission to return true
    vi.mocked(useUserStore).mockReturnValue(true);

    render(<AuthButton perm="user:delete">Delete</AuthButton>);

    const button = screen.getByRole('button', { name: /delete/i });
    expect(button).toBeInTheDocument();
  });

  it('should not render button when user lacks permission', () => {
    // Mock hasPermission to return false
    vi.mocked(useUserStore).mockReturnValue(false);

    render(<AuthButton perm="user:delete">Delete</AuthButton>);

    const button = screen.queryByRole('button', { name: /delete/i });
    expect(button).not.toBeInTheDocument();
  });

  it('should render fallback when user lacks permission and fallback is provided', () => {
    // Mock hasPermission to return false
    vi.mocked(useUserStore).mockReturnValue(false);

    render(
      <AuthButton perm="user:delete" fallback={<span>No permission</span>}>
        Delete
      </AuthButton>
    );

    const fallback = screen.getByText(/no permission/i);
    expect(fallback).toBeInTheDocument();

    const button = screen.queryByRole('button', { name: /delete/i });
    expect(button).not.toBeInTheDocument();
  });

  it('should pass through Button props when user has permission', () => {
    // Mock hasPermission to return true
    vi.mocked(useUserStore).mockReturnValue(true);

    render(
      <AuthButton perm="user:edit" type="primary" danger>
        Edit
      </AuthButton>
    );

    const button = screen.getByRole('button', { name: /edit/i });
    expect(button).toBeInTheDocument();
    expect(button).toHaveClass('ant-btn-primary');
    expect(button).toHaveClass('ant-btn-dangerous');
  });

  it('should handle multiple permission checks correctly', () => {
    // Mock hasPermission to return true for first call, false for second
    let callCount = 0;
    vi.mocked(useUserStore).mockImplementation(() => {
      callCount++;
      return callCount === 1;
    });

    const { rerender } = render(<AuthButton perm="user:create">Create</AuthButton>);
    expect(screen.getByRole('button', { name: /create/i })).toBeInTheDocument();

    rerender(<AuthButton perm="user:delete">Delete</AuthButton>);
    expect(screen.queryByRole('button', { name: /delete/i })).not.toBeInTheDocument();
  });
});
