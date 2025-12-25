<script lang="ts">
  export let text: string;
  export let position: 'top' | 'bottom' | 'left' | 'right' = 'top';

  let tooltipVisible = false;
  let triggerElement: HTMLDivElement;
  let tooltipStyle = '';

  function calculatePosition() {
    if (!triggerElement) return;

    const rect = triggerElement.getBoundingClientRect();
    const gap = 8;

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

    tooltipStyle = `top: ${top}px; left: ${left}px;`;
  }

  function showTooltip() {
    calculatePosition();
    tooltipVisible = true;
  }

  function hideTooltip() {
    tooltipVisible = false;
  }
</script>

<!-- svelte-ignore a11y-no-noninteractive-element-interactions a11y-mouse-events-have-key-events -->
<div class="tooltip-container">
  <div
    bind:this={triggerElement}
    class="tooltip-trigger"
    on:mouseenter={showTooltip}
    on:mouseleave={hideTooltip}
    on:focusin={showTooltip}
    on:focusout={hideTooltip}
    role="tooltip"
    aria-label="Tooltip trigger"
  >
    <slot />
  </div>

  {#if tooltipVisible}
    <div
      class="tooltip fixed z-[9999] px-2 py-1 text-xs rounded bg-gray-900/90 text-white whitespace-nowrap shadow-lg backdrop-blur-sm"
      class:top="{position === 'top'}"
      class:bottom="{position === 'bottom'}"
      class:left="{position === 'left'}"
      class:right="{position === 'right'}"
      style={tooltipStyle}
      role="tooltip"
      aria-label={text}
    >
      {text}
      <div class="tooltip-arrow" role="presentation" />
    </div>
  {/if}
</div>

<style>
  .tooltip-container {
    position: relative;
    display: inline-block;
  }

  .tooltip-trigger {
    display: inline-flex;
  }

  .tooltip {
    pointer-events: none;
    transition: opacity 150ms ease-in-out;
    opacity: 1;
  }

  .tooltip.top {
    transform: translate(-50%, -100%);
  }

  .tooltip.bottom {
    transform: translate(-50%, 0);
  }

  .tooltip.left {
    transform: translate(-100%, -50%);
  }

  .tooltip.right {
    transform: translate(0, -50%);
  }

  .tooltip-arrow {
    position: absolute;
    width: 8px;
    height: 8px;
    background: inherit;
    transform: rotate(45deg);
  }

  .tooltip.top .tooltip-arrow {
    bottom: -4px;
    left: 50%;
    margin-left: -4px;
  }

  .tooltip.bottom .tooltip-arrow {
    top: -4px;
    left: 50%;
    margin-left: -4px;
  }

  .tooltip.left .tooltip-arrow {
    right: -4px;
    top: 50%;
    margin-top: -4px;
  }

  .tooltip.right .tooltip-arrow {
    left: -4px;
    top: 50%;
    margin-top: -4px;
  }
</style>
