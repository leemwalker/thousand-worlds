<script lang="ts">
  // DesktopLayout.svelte
  // Three-column layout for desktop screens
</script>

<div class="flex h-screen bg-gray-900 text-gray-100 overflow-hidden">
  <!-- Left Panel (Map/Navigation) -->
  <div class="w-[250px] bg-gray-800 border-r border-gray-700 flex flex-col shrink-0">
    <div class="p-4 border-b border-gray-700 font-bold text-gray-400 uppercase text-xs tracking-wider">
      Map & Navigation
    </div>
    <div class="flex-1 overflow-y-auto p-4">
      <slot name="left-panel">
        <div class="text-gray-500 text-sm">Map Placeholder</div>
      </slot>
    </div>
  </div>

  <!-- Center Panel (Main Game) -->
  <div class="flex-1 flex flex-col min-w-0">
    <!-- Status Bar -->
    <div class="h-[60px] bg-gray-800 border-b border-gray-700 flex items-center px-4 shrink-0">
      <slot name="status-bar">
        <div class="text-gray-500 text-sm">Status Bar Placeholder</div>
      </slot>
    </div>

    <!-- Main Text Display -->
    <div class="flex-1 overflow-y-auto p-6">
      <slot name="main-display">
        <div class="text-gray-500 text-sm">Game Output Placeholder</div>
      </slot>
    </div>

    <!-- Command Input -->
    <div class="h-[80px] bg-gray-800 border-t border-gray-700 p-4 shrink-0">
      <slot name="command-input">
        <div class="text-gray-500 text-sm">Input Placeholder</div>
      </slot>
    </div>
  </div>

  <!-- Right Panel (Stats/Inventory) -->
  <div class="w-[300px] bg-gray-800 border-l border-gray-700 flex flex-col shrink-0">
    <div class="p-4 border-b border-gray-700 font-bold text-gray-400 uppercase text-xs tracking-wider">
      Character & Inventory
    </div>
    <div class="flex-1 overflow-y-auto p-4">
      <slot name="right-panel">
        <div class="text-gray-500 text-sm">Stats Placeholder</div>
      </slot>
    </div>
  </div>
</div>
