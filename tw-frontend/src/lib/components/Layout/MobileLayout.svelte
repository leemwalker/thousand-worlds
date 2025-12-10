<script lang="ts">
  // MobileLayout.svelte
  // Single column layout optimized for mobile devices
</script>

<div class="flex flex-col h-screen bg-gray-900 text-gray-100 overflow-hidden">
  <!-- Status Bar (Fixed Top) -->
  <div class="h-[60px] bg-gray-800 border-b border-gray-700 flex items-center px-4 shrink-0 z-10">
    <slot name="status-bar">
      <div class="text-gray-500 text-sm">Status Bar Placeholder</div>
    </slot>
  </div>

  <!-- Main Text Display (Scrollable) -->
  <div class="flex-1 overflow-y-auto p-4 relative">
    <slot name="main-display">
      <div class="text-gray-500 text-sm">Game Output Placeholder</div>
    </slot>
    
    <!-- Floating Action Button or Overlay for Map/Inventory could go here -->
  </div>

  <!-- Command Input (Fixed Bottom) -->
  <div class="h-[80px] bg-gray-800 border-t border-gray-700 p-2 shrink-0 z-10">
    <slot name="command-input">
      <div class="text-gray-500 text-sm">Input Placeholder</div>
    </slot>
  </div>

  <!-- Mobile Controls (Optional, e.g. Joystick/Buttons) -->
  <div class="h-[60px] bg-gray-900 border-t border-gray-800 p-2 shrink-0 flex justify-center items-center">
    <slot name="controls">
      <div class="text-gray-600 text-xs">Controls Placeholder</div>
    </slot>
  </div>
</div>
