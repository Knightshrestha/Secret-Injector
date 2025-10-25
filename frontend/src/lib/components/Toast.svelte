<script lang="ts">
  let visible = $state(false);
  let message = $state('');
  let timeout: ReturnType<typeof setTimeout>;

  export function show(msg: string, duration = 3000) {
    message = msg;
    visible = true;

    clearTimeout(timeout);
    timeout = setTimeout(() => {
      visible = false;
    }, duration);
  }
</script>

{#if visible}
  <div class="toast">
    {message}
  </div>
{/if}

<style>
  .toast {
    position: fixed;
    bottom: 2rem;
    right: 2rem;
    background: #323232;
    color: white;
    padding: 1rem 1.5rem;
    border-radius: 8px;
    box-shadow: 0 4px 12px rgba(0, 0, 0, 0.3);
    animation: slideIn 0.3s ease;
    z-index: 1000;
  }

  @keyframes slideIn {
    from {
      transform: translateY(100px);
      opacity: 0;
    }
    to {
      transform: translateY(0);
      opacity: 1;
    }
  }
</style>