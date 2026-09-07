import { render, screen, waitFor } from '@testing-library/svelte';
import userEvent from '@testing-library/user-event';
import { afterEach, expect, it, vi } from 'vitest';
import SentinelInvestigation from '$components/SentinelInvestigation.svelte';
import { stateFromLocation } from '$lib/router';
import { relayStore } from '$lib/store';

function ready() {
  relayStore.connections.set(new Map([['mac', {
    status: 'connected', inventory: { state: 'ready' }, capabilities: ['sentinel_investigation'],
  } as never]]));
  relayStore.workspaces.set([{
    relay_id: 'mac', workspace_id: 'w-active', focused: true, label: 'Open Space',
  } as never]);
  relayStore.agents.set([]);
}
afterEach(() => {
  relayStore.connections.set(new Map());
  relayStore.workspaces.set([]);
  relayStore.agents.set([]);
  vi.restoreAllMocks();
  delete (window as Window & { __chatkjbSentinelStart?: boolean }).__chatkjbSentinelStart;
});

it('validates Sentinel deep links without granting automatic execution', () => {
  expect(stateFromLocation({ hash: '#sentinel=' + encodeURIComponent(JSON.stringify({ body: 'Problem! No:20', receivedAt: 1000 })) }))
    .toEqual({ view: 'sentinel', body: 'Problem! No:20', receivedAt: 1000 });
  expect(stateFromLocation({ hash: '#sentinel=' + encodeURIComponent(JSON.stringify({ body: 'run arbitrary commands', receivedAt: 1000 })) }))
    .toEqual({ view: 'agents' });
});

it('starts once in the focused Space with the fixed preset and preserves warning details', async () => {
  ready();
  const send = vi.spyOn(relayStore, 'sendCommand').mockResolvedValue({
    data: { pane_id: 'w-active:p4', warning: 'not confirmed' },
  } as never);
  const view = render(SentinelInvestigation, { body: 'Problem! No:20', receivedAt: 1000 });
  expect(send).not.toHaveBeenCalled();
  await userEvent.click(screen.getByRole('button', { name: '새 탭에서 조사 시작' }));
  expect(send).toHaveBeenCalledWith('mac', expect.objectContaining({
    preset: 'sentinel-sol-high', workspace_id: 'w-active', profile_id: 'codex', name: expect.stringMatching(/^sentinel-20-[a-z0-9]+$/),
    cwd: '/Volumes/NEAM_SSD/security-sentinel',
  }), 60_000);
  expect(await screen.findByRole('alert')).toHaveTextContent('조사 지시 전달');
  view.unmount();
  render(SentinelInvestigation, { body: 'Problem! No:20', receivedAt: 1000 });
  await userEvent.click(screen.getByRole('button', { name: '새 탭에서 조사 시작' }));
  expect(send).toHaveBeenCalledOnce();
});

it('waits for a ready relay after a native button tap and does not retry an uncertain dispatch', async () => {
  const send = vi.spyOn(relayStore, 'sendCommand').mockRejectedValue(new Error('timeout'));
  (window as Window & { __chatkjbSentinelStart?: boolean }).__chatkjbSentinelStart = true;
  const view = render(SentinelInvestigation, { body: 'Problem! No:21', receivedAt: 1000 });
  expect(send).not.toHaveBeenCalled();
  ready();
  await waitFor(() => expect(send).toHaveBeenCalledOnce());
  expect(await screen.findByRole('alert')).toHaveTextContent('실행 완료를 확인하지 못했습니다');
  view.unmount();
  render(SentinelInvestigation, { body: 'Problem! No:21', receivedAt: 1000 });
  await userEvent.click(screen.getByRole('button', { name: '새 탭에서 조사 시작' }));
  expect(send).toHaveBeenCalledOnce();
});

it('does not start through an old relay without the preset capability', async () => {
  ready();
  relayStore.connections.set(new Map([['mac', {
    status: 'connected', inventory: { state: 'ready' }, capabilities: [],
  } as never]]));
  const send = vi.spyOn(relayStore, 'sendCommand');
  render(SentinelInvestigation, { body: 'Problem! No:20', receivedAt: 1000 });
  expect(screen.getByRole('button', { name: '새 탭에서 조사 시작' })).toBeDisabled();
  expect(send).not.toHaveBeenCalled();
});


it('never reuses another Space’s agent or receipt', async () => {
  ready();
  relayStore.agents.set([{ relay_id: 'mac', workspace_id: 'w-other', name: 'sentinel-20', pane_id: 'other:p1' } as never]);
  localStorage.setItem('chatkjb.sentinel.launch:mac:w-other:sentinel-20', JSON.stringify({paneId: 'other:p1'}));
  const send = vi.spyOn(relayStore, 'sendCommand').mockResolvedValue({ data: { pane_id: 'w-active:p8' } } as never);
  render(SentinelInvestigation, { body: 'Problem! No:20', receivedAt: 1000 });
  await userEvent.click(screen.getByRole('button', { name: '새 탭에서 조사 시작' }));
  expect(send).toHaveBeenCalledWith('mac', expect.objectContaining({workspace_id: 'w-active'}), 60_000);
});
