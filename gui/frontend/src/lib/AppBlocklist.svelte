<script lang="ts">
  import { onMount } from 'svelte';
  import { writable } from 'svelte/store';

  interface BlockedApp {
    name: string;
    exe_path: string;
    commercialName?: string; // Make optional as it will be loaded async
    icon?: string; // Make optional as it will be loaded async
  }
  // This array will hold the names of the currently selected checkboxes.
  // Svelte's `bind:group` directive will automatically keep this array in sync with the UI.
  let selectedApps: string[] = [];
  let blocklistItems = writable<BlockedApp[]>([]);

  async function loadBlocklist(): Promise<void> {
    const res = await fetch('/api/blocklist', { cache: 'no-cache' });
    const data: BlockedApp[] = await res.json();

    if (data && data.length > 0) {
      blocklistItems.set(data);

      // Now, fetch the details for each app
      const detailedItems = await Promise.all(
        data.map(async (app) => {
          if (app.exe_path) {
            const appDetailsRes = await fetch(
              `/api/app-details?path=${encodeURIComponent(app.exe_path)}`
            );
            if (appDetailsRes.ok) {
              const appDetails = await appDetailsRes.json();
              return { ...app, ...appDetails };
            }
          }
          return app; // Return the original app if no path or if fetch fails
        })
      );
      blocklistItems.set(detailedItems);
    } else {
      blocklistItems.set([]);
    }
  }
  let unblockStatus = writable('');

  // Unblocks all applications that are currently selected in the UI.
  async function unblockSelected(): Promise<void> {
    if (selectedApps.length === 0) {
      alert('Vui lòng chọn các ứng dụng để bỏ chặn.');
      return;
    }

    // Create an array of fetch promises, one for each selected app.
    // This allows us to send the removal requests in parallel and handle failures individually.
    const removalPromises = selectedApps.map(async (name) => {
      // The backend /api/unblock endpoint is designed to accept an array of names.
      // Here, we send an array with a single name for each request to get per-item success/failure feedback.
      const response = await fetch('/api/unblock', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ names: [name] }),
      });
      // It's important to check if the request was successful. If not, alert the user.
      if (!response.ok) {
        alert(`Error unblocking ${name}: ${response.statusText}`);
        throw new Error(`Failed to unblock ${name}`);
      }
    });

    try {
      // Wait for all the removal requests to complete.
      await Promise.all(removalPromises);
    } catch {
      // If any of the promises fail, the error will be caught here.
      // The individual errors are already alerted, so we just stop execution.
      return;
    }

    unblockStatus.set('Đã bỏ chặn: ' + selectedApps.join(', '));
    setTimeout(() => {
      unblockStatus.set('');
    }, 3000);
    loadBlocklist(); // Refresh the list
    selectedApps = []; // Clear the selection
  }

  async function clearBlocklist(): Promise<void> {
    if (confirm('Bạn có chắc chắn muốn xóa toàn bộ danh sách chặn không?')) {
      await fetch('/api/blocklist/clear', { method: 'POST' });
      unblockStatus.set('Đã xóa toàn bộ danh sách chặn.');
      setTimeout(() => {
        unblockStatus.set('');
      }, 3000);
      loadBlocklist(); // Refresh the list
    }
  }

  async function saveBlocklist(): Promise<void> {
    const response = await fetch('/api/blocklist/save');
    const blob = await response.blob();
    const url = window.URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.style.display = 'none';
    a.href = url;
    a.download = 'procguard_blocklist.json';
    document.body.appendChild(a);
    a.click();
    window.URL.revokeObjectURL(url);
  }

  async function loadBlocklistFile(event: Event): Promise<void> {
    const file = (event.target as HTMLInputElement).files?.[0];
    if (!file) {
      return;
    }
    const formData = new FormData();
    formData.append('file', file);

    await fetch('/api/blocklist/load', {
      method: 'POST',
      body: formData,
    });

    unblockStatus.set('Đã tải lên và hợp nhất danh sách chặn.');
    setTimeout(() => {
      unblockStatus.set('');
    }, 3000);
    loadBlocklist(); // Refresh the list
  }

  onMount(() => {
    loadBlocklist();
  });
</script>

<div class="card mt-3">
  <div class="card-body">
    <h5 class="card-title">Các ứng dụng bị chặn</h5>
    <div class="btn-toolbar" role="toolbar">
      <div class="btn-group me-2" role="group">
        <button class="btn btn-primary" on:click={unblockSelected}>
          Bỏ chặn mục đã chọn
        </button>
        <button type="button" class="btn btn-danger" on:click={clearBlocklist}>
          Xóa toàn bộ
        </button>
      </div>
      <div class="btn-group" role="group">
        <button
          type="button"
          class="btn btn-outline-secondary"
          on:click={saveBlocklist}
        >
          Lưu danh sách
        </button>
        <button
          type="button"
          class="btn btn-outline-secondary"
          on:click={() => document.getElementById('load-input')?.click()}
        >
          Tải lên danh sách
        </button>
      </div>
    </div>
    <input
      type="file"
      id="load-input"
      style="display: none"
      on:change={loadBlocklistFile}
    />
    {#if $unblockStatus}
      <span id="unblock-status" class="form-text">{$unblockStatus}</span>
    {/if}
    <div id="blocklist-items" class="list-group mt-3">
      {#if $blocklistItems.length > 0}
        {#each $blocklistItems as app (app.name)}
          <label class="list-group-item d-flex align-items-center">
            <input
              class="form-check-input me-2"
              type="checkbox"
              name="blocked-app"
              value={app.name}
              bind:group={selectedApps}
            />
            {#if app.icon}
              <img
                src="data:image/png;base64,{app.icon}"
                class="me-2"
                style="width: 24px; height: 24px;"
                alt="App Icon"
              />
            {:else}
              <div class="me-2" style="width: 24px; height: 24px;"></div>
            {/if}
            <span class="fw-bold me-2">{app.commercialName || app.name}</span>
          </label>
        {/each}
      {:else}
        <div class="list-group-item">Hiện không có ứng dụng nào bị chặn.</div>
      {/if}
    </div>
  </div>
</div>
