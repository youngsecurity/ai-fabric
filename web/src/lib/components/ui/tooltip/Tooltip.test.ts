import { describe, it, expect } from 'vitest';

describe('Tooltip positioning logic', () => {
  const gap = 8;

  function calculatePosition(
    rect: { top: number; bottom: number; left: number; right: number; width: number; height: number },
    position: 'top' | 'bottom' | 'left' | 'right'
  ) {
    let top = 0;
    let left = 0;

    switch (position) {
      case 'top':
        top = rect.top - gap;
        left = rect.left + rect.width / 2;
        break;
      case 'bottom':
        top = rect.bottom + gap;
        left = rect.left + rect.width / 2;
        break;
      case 'left':
        top = rect.top + rect.height / 2;
        left = rect.left - gap;
        break;
      case 'right':
        top = rect.top + rect.height / 2;
        left = rect.right + gap;
        break;
    }

    return { top, left };
  }

  const mockRect = {
    top: 100,
    bottom: 130,
    left: 200,
    right: 300,
    width: 100,
    height: 30
  };

  it('calculates top position correctly', () => {
    const result = calculatePosition(mockRect, 'top');
    expect(result.top).toBe(92); // 100 - 8
    expect(result.left).toBe(250); // 200 + 100/2
  });

  it('calculates bottom position correctly', () => {
    const result = calculatePosition(mockRect, 'bottom');
    expect(result.top).toBe(138); // 130 + 8
    expect(result.left).toBe(250); // 200 + 100/2
  });

  it('calculates left position correctly', () => {
    const result = calculatePosition(mockRect, 'left');
    expect(result.top).toBe(115); // 100 + 30/2
    expect(result.left).toBe(192); // 200 - 8
  });

  it('calculates right position correctly', () => {
    const result = calculatePosition(mockRect, 'right');
    expect(result.top).toBe(115); // 100 + 30/2
    expect(result.left).toBe(308); // 300 + 8
  });
});
