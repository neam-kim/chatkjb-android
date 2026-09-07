<script lang="ts">
  import { onMount } from 'svelte';
  import Button from '$components/ui/Button.svelte';
  import Card from '$components/ui/Card.svelte';
  import { clientPaneId } from '$lib/agents';
  import { replaceView } from '$lib/router';
  import { relayStore } from '$lib/store';

  let { body, receivedAt }: { body: string; receivedAt: number } = $props();
  const connections = relayStore.connections;
  const workspaces = relayStore.workspaces;
  const agents = relayStore.agents;
  let authorized = $state(false);
  let attempted = $state(false);
  let busy = $state(false);
  let status = $state('Mac mini 연결을 기다리고 있습니다.');
  let paneId = $state('');
  let failed = $state(false);
  const legacyName = $derived(`sentinel-${body.replace('Problem! No:', '')}`);
  const targets = $derived($workspaces.filter((workspace) => {
    const connection = $connections.get(workspace.relay_id);
    return workspace.focused && connection?.status === 'connected'
      && connection.inventory.state === 'ready'
      && connection.capabilities.includes('sentinel_investigation');
  }));
  const target = $derived(targets.length === 1 ? targets[0] : null);

  onMount(() => {
    const bridge = window as Window & { __chatkjbSentinelStart?: boolean };
    const authorize = () => {
      if (!bridge.__chatkjbSentinelStart) return;
      delete bridge.__chatkjbSentinelStart;
      authorized = true;
    };
    window.addEventListener('chatkjb-sentinel-start', authorize);
    authorize();
    return () => window.removeEventListener('chatkjb-sentinel-start', authorize);
  });

  $effect(() => {
    if (authorized && target && !attempted) void start();
  });

  async function start() {
    if (!target || busy || attempted) return;
    const workspace = target;
    attempted = true;
    let hash = 2166136261;
    for (const char of workspace.workspace_id) hash = Math.imul(hash ^ char.charCodeAt(0), 16777619);
    const name = `${legacyName}-${(hash >>> 0).toString(36)}`;
    const existing = $agents.find((agent) => agent.relay_id === workspace.relay_id
      && agent.workspace_id === workspace.workspace_id
      && (agent.name === name || agent.name === legacyName));
    if (existing) {
      paneId = existing.pane_id;
      status = '이 알림의 조사 탭이 이미 열려 있습니다.';
      return;
    }
    const key = `chatkjb.sentinel.launch:${workspace.relay_id}:${workspace.workspace_id}:${legacyName}`;
    try {
      const receipt = localStorage.getItem(key);
      if (receipt) {
        const saved = JSON.parse(receipt);
        paneId = String(saved.paneId || '');
        status = paneId ? '조사 요청을 이미 보냈습니다.' : '이 알림의 실행 요청이 이미 전송되었습니다. Herdr 탭에서 실행 상태를 확인해 주세요.';
        return;
      }
      // Persist before dispatch: a timeout, back navigation, or rotation must
      // never create a second agent for the same problem automatically.
      localStorage.setItem(key, JSON.stringify({ requestedAt: Date.now() }));
    } catch {
      failed = true;
      status = '중복 실행을 방지할 기록을 저장하지 못했습니다. 저장 공간을 확인해 주세요.';
      return;
    }
    busy = true;
    status = `${workspace.label || '현재 Space'}에 조사 탭을 여는 중…`;
    try {
      const result = await relayStore.sendCommand(workspace.relay_id, {
        type: 'agent_start', profile_id: 'codex', preset: 'sentinel-sol-high',
        name, cwd: '/Volumes/NEAM_SSD/security-sentinel', workspace_id: workspace.workspace_id,
        prompt: `Sentinel 알림 ${body}의 원인을 조사하세요. 휴대폰 수신 시각: ${new Date(receivedAt).toISOString()} (발생 시각과 다를 수 있음). `
          + 'sentinel-briefing과 관련 로컬 상태 및 로그를 읽기 전용으로 확인하고, 이 번호의 근거와 현재 복구 여부를 구분해 설명하세요. '
          + '최신 상태가 다른 번호이면 그것을 이 알림의 원인으로 단정하지 마세요. '
          + '인증정보, raw 환경변수, 알림 endpoint/topic/token, 원문 인증 로그를 출력하지 마세요. '
          + '설정 변경, 서비스 재시작, 알림 발송 또는 baseline 초기화는 하지 말고 필요한 후속 조치를 보고하세요.',
      }, 60_000);
      const rawPaneId = String(result.data?.pane_id || '');
      paneId = rawPaneId ? clientPaneId(workspace.relay_id, rawPaneId) : '';
      const warning = String(result.data?.warning || '');
      failed = Boolean(warning);
      status = warning ? '탭은 열렸지만 조사 지시 전달을 확인하지 못했습니다. 탭에서 확인해 주세요.' : 'Codex Sol / high에 원인 조사를 요청했습니다.';
      localStorage.setItem(key, JSON.stringify({ paneId, requestedAt: Date.now() }));
    } catch {
      failed = true;
      status = '실행 완료를 확인하지 못했습니다. 중복 실행을 피하도록 Herdr 탭에서 상태를 확인해 주세요.';
    } finally {
      busy = false;
    }
  }
</script>

<main class="page" aria-labelledby="sentinel-title">
  <h2 id="sentinel-title">Sentinel 원인 조사</h2>
  <Card>
    <p>{body}</p>
    <p>수신 {new Date(receivedAt).toLocaleString()}</p>
    <p>현재 열린 Space의 새 탭 · Codex Sol / high</p>
    {#if !attempted && !target}
      <p role="status">{targets.length > 1 ? '활성 Space를 하나로 선택해 주세요.' : '연결된 Herdr와 열린 Space가 필요합니다. 연결 상태와 relay 업데이트를 확인해 주세요.'}</p>
    {:else}
      <p role={failed ? 'alert' : 'status'}>{status}</p>
    {/if}
    {#if !authorized && !attempted}
      <Button disabled={!target || busy} onclick={start}>새 탭에서 조사 시작</Button>
    {/if}
    {#if paneId}<Button onclick={() => replaceView({ view: 'terminal', paneId })}>조사 탭 열기</Button>{/if}
    <Button variant="secondary" onclick={() => replaceView({ view: 'agents' })}>Herdr 탭 목록</Button>
  </Card>
</main>
