import { useEffect, useMemo, useRef, useState } from 'react';
import type { ReactNode } from 'react';
import { AlertTriangle, Bot, Check, History, Loader2, MessageSquare, Plus, RotateCcw, Send, ShieldCheck, Trash2, X } from 'lucide-react';
import { toast } from 'sonner';
import { api } from './api';
import { pluginWebSocketURL } from './plugin-paths';
import type { AgentInfo, ChatEvent, ChatExecutionEvent, ChatInteractionInfo, ChatInteractionResponseScope, ChatMessageInfo, ChatSessionInfo } from './types';
import { Alert, Badge, Button, ChatComposer, ChatMessage, ChatThread, CheckboxField, ConfirmDialog, EmptyState, IconButton, Input, Textarea } from '@/components/ui';
import { cn, normalizeWorkspacePath, workspaceRootPath } from '@/lib/utils';

type RenderedChatMessage = ChatMessageInfo & {
  renderSequence: number;
  executionEvents?: ChatExecutionEvent[];
  terminalId?: string;
  transient?: boolean;
};

type ChatSnapshot = {
  session: ChatSessionInfo | null;
  messages: RenderedChatMessage[];
  interactions: ChatInteractionInfo[];
  interactionDrafts: Record<string, string>;
  input: string;
  status: string;
  error: string;
  notice: string;
  renderSequence: number;
};

type SelectSessionOptions = {
  notice?: string;
};

type Props = {
  active: boolean;
  chatHistoryOpen: boolean;
  connected: boolean;
  agent?: AgentInfo;
  preferredSessionId?: string;
  workspacePath: string;
  onChatStateChange?: (state: { activeChatSessionId?: string; chatHistoryOpen?: boolean }) => void;
  onAgentWorkingChange?: (state?: 'working' | 'waiting_input') => void;
};

let transientMessageId = -1;
const chatSnapshots = new Map<string, ChatSnapshot>();
const interactionProgressTerminalId = 'interaction-continuation-progress';

export default function ChatPanel({ active, chatHistoryOpen, connected, agent, preferredSessionId, workspacePath, onAgentWorkingChange, onChatStateChange }: Props) {
  const [sessions, setSessions] = useState<ChatSessionInfo[]>([]);
  const [session, setSession] = useState<ChatSessionInfo | null>(null);
  const [messages, setMessages] = useState<RenderedChatMessage[]>([]);
  const [interactions, setInteractions] = useState<ChatInteractionInfo[]>([]);
  const [interactionDrafts, setInteractionDrafts] = useState<Record<string, string>>({});
  const [interactionSubmitting, setInteractionSubmitting] = useState<Record<string, boolean>>({});
  const [input, setInput] = useState('');
  const [loading, setLoading] = useState(false);
  const [sending, setSending] = useState(false);
  const [socketReady, setSocketReady] = useState(false);
  const [status, setStatus] = useState('idle');
  const [error, setError] = useState('');
  const [notice, setNotice] = useState('');
  const [deleteTarget, setDeleteTarget] = useState<ChatSessionInfo | null>(null);
  const socketRef = useRef<WebSocket | null>(null);
  const messagesRef = useRef<HTMLDivElement | null>(null);
  const renderSequenceRef = useRef(0);
  const sessionRef = useRef<ChatSessionInfo | null>(null);
  const messagesStateRef = useRef<RenderedChatMessage[]>([]);
  const interactionsRef = useRef<ChatInteractionInfo[]>([]);
  const interactionDraftsRef = useRef<Record<string, string>>({});
  const inputRef = useRef('');
  const statusRef = useRef('idle');
  const errorRef = useRef('');
  const noticeRef = useRef('');
  const running = agent?.runtimeStatus === 'running';
  const effectiveWorkspacePath = normalizeWorkspacePath(workspacePath || workspaceRootPath);
  const canRecover = Boolean(session && status === 'error');
  const activeSessionId = session?.id ?? '';
  const pendingInteractions = useMemo(
    () => [...interactions].filter((item) => item.status === 'pending').sort((a, b) => a.createdAt - b.createdAt),
    [interactions],
  );
  const firstPendingInteractionId = pendingInteractions[0]?.id ?? '';

  const orderedMessages = useMemo(
    () => [...messages].sort((a, b) => a.renderSequence - b.renderSequence),
    [messages],
  );
  const hasStreamingAssistant = orderedMessages.some((message) => (
    message.role === 'assistant' && message.status === 'streaming'
  ));
  const activeSessionWorking = session ? isProcessingSession(session) : false;
  const waitingInput = status === 'waiting_input' || pendingInteractions.length > 0 || session?.runtimeState === 'waiting_input';
  const processing = running && !['closed', 'error', 'exited', 'offline'].includes(status) && !waitingInput && (sending || hasStreamingAssistant || activeSessionWorking);
  const sendBlocked = !running || !socketReady || waitingInput;
  const canSend = !sendBlocked && input.trim().length > 0;
  const anySessionWaiting = running && (waitingInput || sessions.some(isWaitingSession));
  const anySessionWorking = running && (processing || sessions.some(isProcessingSession));

  useEffect(() => {
    let cancelled = false;
    if (!connected || !agent) {
      if (sessionRef.current?.id) {
        persistCurrentSnapshot(sessionRef.current.id);
      }
      setSessions([]);
      setSession(null);
      setMessages([]);
      setInteractions([]);
      setInteractionDrafts({});
      setInteractionSubmitting({});
      setSending(false);
      setSocketReady(false);
      setInput('');
      setNotice('');
      disconnect();
      return () => {
        cancelled = true;
      };
    }
    if (sessionRef.current?.id) {
      persistCurrentSnapshot(sessionRef.current.id);
    }
    disconnect();
    setSending(false);
    setSocketReady(false);
    void initializeChat(agent, () => !cancelled);
    return () => {
      cancelled = true;
      if (sessionRef.current?.id) {
        persistCurrentSnapshot(sessionRef.current.id);
      }
      disconnect();
    };
  }, [connected, agent?.id, agent?.runtimeStatus, effectiveWorkspacePath]);

  useEffect(() => {
    sessionRef.current = session;
  }, [session]);

  useEffect(() => {
    messagesStateRef.current = messages;
  }, [messages]);

  useEffect(() => {
    interactionsRef.current = interactions;
  }, [interactions]);

  useEffect(() => {
    interactionDraftsRef.current = interactionDrafts;
  }, [interactionDrafts]);

  useEffect(() => {
    inputRef.current = input;
  }, [input]);

  useEffect(() => {
    statusRef.current = status;
  }, [status]);

  useEffect(() => {
    errorRef.current = error;
  }, [error]);

  useEffect(() => {
    noticeRef.current = notice;
  }, [notice]);

  useEffect(() => {
    if (session?.id) {
      persistCurrentSnapshot(session.id);
    }
  }, [error, input, interactionDrafts, interactions, messages, notice, session, status]);

  useEffect(() => {
    onAgentWorkingChange?.(anySessionWaiting ? 'waiting_input' : anySessionWorking ? 'working' : undefined);
  }, [anySessionWaiting, anySessionWorking, onAgentWorkingChange]);

  useEffect(() => {
    messagesRef.current?.scrollTo({ top: messagesRef.current.scrollHeight });
  }, [messages]);

  function persistCurrentSnapshot(sessionId: string) {
    persistSnapshot(sessionId, {
      session: sessionRef.current,
      messages: messagesStateRef.current,
      interactions: interactionsRef.current,
      interactionDrafts: interactionDraftsRef.current,
      input: inputRef.current,
      status: statusRef.current,
      error: errorRef.current,
      notice: noticeRef.current,
      renderSequence: renderSequenceRef.current,
    });
  }

  async function initializeChat(target: AgentInfo, shouldApply: () => boolean) {
    if (shouldApply()) {
      setLoading(true);
      setError('');
      setNotice('');
    }
    try {
      let nextSessions = await api.listAgentChatSessions(target.id);
      let nextSession = nextSessions.find((item) => item.id === preferredSessionId) ?? nextSessions[0];
      if (!nextSession) {
        nextSession = await api.createAgentChatSession(target.id);
        nextSessions = [nextSession];
      }
      if (!shouldApply()) {
        return;
      }
      setSessions(nextSessions);
      onChatStateChange?.({ activeChatSessionId: nextSession.id, chatHistoryOpen });
      await loadChat(target.id, nextSession.id, shouldApply);
      if (shouldApply()) {
        connect(target, nextSession.id);
      }
    } catch (err) {
      if (shouldApply()) {
        setError((err as Error).message);
      }
    } finally {
      if (shouldApply()) {
        setLoading(false);
      }
    }
  }

  async function loadChat(agentId: string, sessionId: string, shouldApply: () => boolean) {
    if (shouldApply()) {
      setLoading(true);
      setError('');
    }
    try {
      const response = await api.getAgentChatMessages(agentId, sessionId);
      const loadedMessages = (response.messages ?? []).map((message) => ({
        ...message,
        renderSequence: message.sequence,
        executionEvents: [],
      }));
      const cached = chatSnapshots.get(sessionId);
      const [nextMessages, loadedInteractions] = await Promise.all([
        Promise.resolve(mergeLoadedMessages(loadedMessages, cached?.messages ?? [])),
        api.listAgentChatInteractions(agentId, sessionId, { status: 'pending' }).catch(() => []),
      ]);
      const nextRenderSequence = nextMessages.reduce((max, message) => Math.max(max, message.renderSequence), 0);
      const nextInteractions = mergeInteractions(cached?.interactions ?? [], loadedInteractions);
      const pendingIds = new Set(nextInteractions.filter((item) => item.status === 'pending').map((item) => item.id));
      const nextDrafts = Object.fromEntries(Object.entries(cached?.interactionDrafts ?? {}).filter(([interactionId]) => pendingIds.has(interactionId)));
      const nextSnapshot: ChatSnapshot = {
        session: response.session,
        messages: nextMessages,
        interactions: nextInteractions,
        interactionDrafts: nextDrafts,
        input: cached?.input ?? '',
        status: nextInteractions.some((item) => item.status === 'pending') ? 'waiting_input' : response.session.runtimeState || response.session.status || 'idle',
        error: '',
        notice: response.session.lastError || '',
        renderSequence: nextRenderSequence,
      };
      chatSnapshots.set(sessionId, nextSnapshot);
      if (!shouldApply()) {
        return;
      }
      setSessions((current) => upsertChatSession(current, response.session));
      setSession(nextSnapshot.session);
      renderSequenceRef.current = nextSnapshot.renderSequence;
      setMessages(nextSnapshot.messages);
      setInteractions(nextSnapshot.interactions);
      setInteractionDrafts(nextSnapshot.interactionDrafts);
      setInput(nextSnapshot.input);
      setStatus(nextSnapshot.status);
      setNotice(nextSnapshot.notice);
    } catch (err) {
      if (shouldApply()) {
        setError((err as Error).message);
      }
    } finally {
      if (shouldApply()) {
        setLoading(false);
      }
    }
  }

  function connect(target: AgentInfo, sessionId: string) {
    disconnect();
    if (target.runtimeStatus !== 'running' || !sessionId) {
      return;
    }
    const params = new URLSearchParams({ cwd: effectiveWorkspacePath });
    const socket = new WebSocket(
      pluginWebSocketURL(
        `/agents/${encodeURIComponent(target.id)}/chat/sessions/${encodeURIComponent(sessionId)}?${params}`,
      ),
    );
    socketRef.current = socket;
    setStatus('connecting');
    setSocketReady(false);
    socket.addEventListener('open', () => {
      setSocketReady(true);
      setStatus('connected');
    });
    socket.addEventListener('message', (event) => {
      if (socketRef.current === socket) {
        const parsed = parseIncomingChatEvent(event.data);
        if (parsed) {
          handleEvent(parsed);
          return;
        }
        appendRawAssistantOutput(readIncomingText(event.data));
      }
    });
    socket.addEventListener('close', () => {
      if (socketRef.current === socket) {
        socketRef.current = null;
        setSocketReady(false);
        setStatus(target.runtimeStatus === 'running' ? 'closed' : 'offline');
      }
    });
    socket.addEventListener('error', () => {
      setSocketReady(false);
      setError('对话连接失败');
    });
  }

  function disconnect() {
    socketRef.current?.close();
    socketRef.current = null;
    setSocketReady(false);
  }

  function nextRenderSequence() {
    renderSequenceRef.current += 1;
    return renderSequenceRef.current;
  }

  function upsertMessage(message: ChatMessageInfo) {
    setMessages((current) => {
      const index = current.findIndex((item) => item.id === message.id);
      if (index === -1) {
        return [...current, { ...message, renderSequence: nextRenderSequence(), executionEvents: defaultExecutionEvents(message) }];
      }
      const next = [...current];
      next[index] = { ...message, renderSequence: next[index].renderSequence, executionEvents: next[index].executionEvents ?? [] };
      return next;
    });
  }

  function appendMessageContent(id: number, content: string) {
    setMessages((current) => current.map((message) => (message.id === id ? { ...message, content: message.content + content } : message)));
  }

  function appendTerminalOutput(terminalId: string, content: string) {
    if (!content) {
      return;
    }
    setMessages((current) => {
      const latest = [...current].sort((a, b) => b.renderSequence - a.renderSequence)[0];
      if (latest?.role === 'terminal' && latest.terminalId === terminalId) {
        return current.map((item) => item.id === latest.id ? { ...item, content: item.content + content, status: 'streaming' } : item);
      }
      const order = nextRenderSequence();
      const next: RenderedChatMessage = {
        id: transientMessageId,
        sessionId: session?.id || '',
        sequence: order,
        renderSequence: order,
        role: 'terminal',
        content,
        status: 'streaming',
        metadata: '{}',
        createdAt: Date.now(),
        updatedAt: Date.now(),
        terminalId,
        transient: true,
      };
      transientMessageId -= 1;
      return [...current, next];
    });
  }

  function appendRawAssistantOutput(content: string) {
    if (!content) {
      return;
    }
    clearInteractionProgressMessage();
    setMessages((current) => {
      const latest = [...current].sort((a, b) => b.renderSequence - a.renderSequence)[0];
      if (latest?.role === 'assistant' && latest.transient && latest.terminalId === 'raw-chat-output') {
        return current.map((item) => item.id === latest.id ? { ...item, content: item.content + content, status: 'complete' } : item);
      }
      const order = nextRenderSequence();
      const next: RenderedChatMessage = {
        id: transientMessageId,
        sessionId: session?.id || '',
        sequence: order,
        renderSequence: order,
        role: 'assistant',
        content,
        status: 'complete',
        metadata: '{}',
        createdAt: Date.now(),
        updatedAt: Date.now(),
        terminalId: 'raw-chat-output',
        transient: true,
      };
      transientMessageId -= 1;
      return [...current, next];
    });
    setSending(false);
    markCurrentSessionIdle();
  }

  function appendExecutionEvent(messageId: number, event: ChatExecutionEvent) {
    const safeEvent = normalizeExecutionEvent(event);
    if (!safeEvent) {
      return;
    }
    clearInteractionProgressMessage();
    setMessages((current) => current.map((message) => {
      if (message.id !== messageId) {
        return message;
      }
      const existing = message.executionEvents ?? [];
      const nextEvents = mergeExecutionEvent(existing, safeEvent);
      return { ...message, executionEvents: nextEvents.slice(-24) };
    }));
  }

  function upsertInteraction(interaction: ChatInteractionInfo) {
    const nextInteractions = mergeInteractions(interactionsRef.current, [interaction]);
    const sessionHasPending = nextInteractions.some((item) => item.sessionId === interaction.sessionId && item.status === 'pending');
    interactionsRef.current = nextInteractions;
    setInteractions(nextInteractions);
    setSession((current) => {
      if (!current || current.id !== interaction.sessionId) {
        return current;
      }
      if (sessionHasPending) {
        return { ...current, status: 'waiting_input', runtimeState: 'waiting_input', lastActiveAt: Date.now() };
      }
      if (isWaitingSession(current)) {
        return { ...current, status: 'running', runtimeState: 'running', lastActiveAt: Date.now() };
      }
      return current;
    });
    setSessions((current) => current.map((item) => (
      item.id === interaction.sessionId && sessionHasPending
        ? { ...item, status: 'waiting_input', runtimeState: 'waiting_input', lastActiveAt: Date.now() }
        : item.id === interaction.sessionId && isWaitingSession(item)
        ? { ...item, status: 'running', runtimeState: 'running', lastActiveAt: Date.now() }
        : item
    )));
  }

  function hasPendingInteraction() {
    return interactionsRef.current.some((item) => item.status === 'pending');
  }

  function appendInteractionProgressMessage(interaction: ChatInteractionInfo) {
    if (interactionsRef.current.some((item) => item.sessionId === interaction.sessionId && item.status === 'pending')) {
      return;
    }
    setMessages((current) => {
      const next = current.filter((message) => !isInteractionProgressMessage(message));
      const order = nextRenderSequence();
      return [
        ...next,
        {
          id: transientMessageId,
          sessionId: interaction.sessionId || sessionRef.current?.id || '',
          sequence: order,
          renderSequence: order,
          role: 'assistant',
          content: '响应已提交，Agent 正在继续执行。',
          status: 'streaming',
          metadata: '{}',
          createdAt: Date.now(),
          updatedAt: Date.now(),
          terminalId: interactionProgressTerminalId,
          transient: true,
          executionEvents: [{
            id: `local-${interaction.id}-continuation`,
            kind: 'status',
            title: '响应已提交',
            detail: 'Agent 正在继续执行',
            status: 'running',
            createdAt: new Date().toISOString(),
          }],
        },
      ];
    });
    transientMessageId -= 1;
  }

  function clearInteractionProgressMessage() {
    setMessages((current) => {
      if (!current.some(isInteractionProgressMessage)) {
        return current;
      }
      return current.filter((message) => !isInteractionProgressMessage(message));
    });
  }

  function clearEmptyStreamingAssistantPlaceholders() {
    setMessages((current) => {
      if (!current.some(isEmptyStreamingAssistantPlaceholder)) {
        return current;
      }
      return current.filter((message) => !isEmptyStreamingAssistantPlaceholder(message));
    });
  }

  function handleEvent(event: ChatEvent) {
    if (event.session) {
      setSession(event.session);
      setSessions((current) => upsertChatSession(current, event.session!));
    }
    if (event.status) {
      setStatus(event.status);
      if (!isProcessingStatus(event.status)) {
        setSending(false);
      }
    }
    if (event.notice) setNotice(event.notice);
    if (event.error) setError(event.error);
    if (event.interaction) {
      upsertInteraction(event.interaction);
      if (event.interaction.status === 'pending') {
        setStatus('waiting_input');
        setSending(false);
      } else if (event.type === 'interaction_resolved') {
        setStatus('running');
      }
    }

    if (event.type === 'message' && event.message) {
      if (event.message.role === 'assistant') {
        clearInteractionProgressMessage();
      }
      upsertMessage(event.message);
      touchSessionFromMessage(event.message);
      if (event.message.role === 'assistant' && event.message.status === 'streaming') setSending(true);
    }
    if (event.type === 'assistant_delta' && event.messageId && event.content) {
      clearInteractionProgressMessage();
      appendMessageContent(event.messageId, event.content);
      setSending(true);
    }
    if (event.type === 'execution_event' && event.messageId && event.event) {
      appendExecutionEvent(event.messageId, event.event);
      setSending(true);
    }
    if (event.type === 'terminal_output') {
      clearInteractionProgressMessage();
      appendTerminalOutput(event.terminalId || 'default', event.content || '');
    }
    if (event.type === 'message_complete') {
      clearInteractionProgressMessage();
      clearEmptyStreamingAssistantPlaceholders();
      if (event.message) upsertMessage(event.message);
      setSending(false);
      if (!hasPendingInteraction()) {
        markCurrentSessionIdle();
      }
    }
    if (event.type === 'recoverable_error' || event.type === 'error' || event.type === 'resume_required') {
      clearInteractionProgressMessage();
      clearEmptyStreamingAssistantPlaceholders();
      setSending(false);
      if (event.type === 'recoverable_error') setStatus('error');
      if (event.type === 'resume_required') setStatus('exited');
    }
  }

  function sendMessage() {
    const content = input.trim();
    if (!canSend || !socketRef.current || socketRef.current.readyState !== WebSocket.OPEN) {
      return;
    }
    setError('');
    setNotice('');
    setSending(true);
    socketRef.current.send(JSON.stringify({ type: 'user_message', content }));
    setInput('');
  }

  function markCurrentSessionIdle() {
    setStatus('idle');
    setSession((current) => current ? { ...current, status: 'idle', runtimeState: 'idle', lastError: '' } : current);
    setSessions((current) => current.map((item) => (
      item.id === sessionRef.current?.id ? { ...item, status: 'idle', runtimeState: 'idle', lastError: '' } : item
    )));
  }

  async function submitInteraction(interaction: ChatInteractionInfo, response: Record<string, unknown>, responseMode: string, responseScope: ChatInteractionResponseScope = '') {
    if (!agent || !session) {
      return;
    }
    setInteractionSubmitting((current) => ({ ...current, [interaction.id]: true }));
    setError('');
    try {
      const updated = await api.updateAgentChatInteractionResponse(agent.id, session.id, interaction.id, {
        response: JSON.stringify(response),
        responseMode,
        responseScope,
      });
      upsertInteraction(updated);
      setInteractionDrafts((current) => {
        const next = { ...current };
        delete next[interaction.id];
        return next;
      });
      if (!interactionsRef.current.some((item) => item.status === 'pending')) {
        setStatus('running');
        setSending(true);
        if (updated.status === 'resolved') {
          appendInteractionProgressMessage(updated);
        }
      }
    } catch (err) {
      setError((err as Error).message);
    } finally {
      setInteractionSubmitting((current) => {
        const next = { ...current };
        delete next[interaction.id];
        return next;
      });
    }
  }

  async function cancelInteraction(interaction: ChatInteractionInfo) {
    if (!agent || !session) {
      return;
    }
    setInteractionSubmitting((current) => ({ ...current, [interaction.id]: true }));
    setError('');
    try {
      const updated = await api.updateAgentChatInteractionStatus(agent.id, session.id, interaction.id, { status: 'cancelled' });
      upsertInteraction(updated);
      setInteractionDrafts((current) => {
        const next = { ...current };
        delete next[interaction.id];
        return next;
      });
      setSending(false);
      if (!interactionsRef.current.some((item) => item.id !== interaction.id && item.status === 'pending')) {
        markCurrentSessionIdle();
      }
    } catch (err) {
      setError((err as Error).message);
    } finally {
      setInteractionSubmitting((current) => {
        const next = { ...current };
        delete next[interaction.id];
        return next;
      });
    }
  }

  async function recoverChat() {
    if (!agent || !session) {
      return;
    }
    setLoading(true);
    setError('');
    try {
      const response = await api.recoverAgentChat(agent.id, session.id);
      if (response.session) {
        setSession(response.session);
        setSessions((current) => upsertChatSession(current, response.session!));
        setStatus(response.session.runtimeState);
      }
      if (response.message) {
        upsertMessage(response.message);
      }
      setNotice('已根据历史记录启动新的对话进程。');
    } catch (err) {
      setError((err as Error).message);
    } finally {
      setLoading(false);
    }
  }

  async function createSession() {
    if (!agent) {
      return;
    }
    if (session?.id) {
      persistCurrentSnapshot(session.id);
    }
    const reusable = findReusableEmptySession(sessions, session?.id);
    if (reusable) {
      if (reusable.id === session?.id) {
        setError('');
        setNotice('已切换到现有空对话。');
        return;
      }
      await selectSession(reusable, { notice: '已切换到现有空对话。' });
      return;
    }
    setLoading(true);
    setError('');
    setNotice('');
    disconnect();
    try {
      const created = await api.createAgentChatSession(agent.id);
      setSessions((current) => upsertChatSession([created, ...current], created));
      applyEmptySession(created);
      onChatStateChange?.({ activeChatSessionId: created.id, chatHistoryOpen });
      if (running) {
        connect(agent, created.id);
      }
    } catch (err) {
      setError((err as Error).message);
    } finally {
      setLoading(false);
    }
  }

  async function deleteSession(target: ChatSessionInfo) {
    if (!agent) {
      return;
    }
    setLoading(true);
    setError('');
    setNotice('');
    try {
      await api.deleteAgentChatSession(agent.id, target.id);
      chatSnapshots.delete(target.id);
      setDeleteTarget(null);
      const remaining = sessions.filter((item) => item.id !== target.id);
      setSessions(remaining);
      if (target.id !== session?.id) {
        showChatDeleteToast();
        return;
      }
      disconnect();
      const fallback = remaining[0] ?? await api.createAgentChatSession(agent.id);
      const nextSessions = remaining.length ? remaining : [fallback];
      setSessions(nextSessions);
      onChatStateChange?.({ activeChatSessionId: fallback.id, chatHistoryOpen });
      await loadChat(agent.id, fallback.id, () => true);
      if (running) {
        connect(agent, fallback.id);
      }
      showChatDeleteToast();
    } catch (err) {
      setError((err as Error).message);
    } finally {
      setLoading(false);
    }
  }

  async function selectSession(nextSession: ChatSessionInfo, options: SelectSessionOptions = {}) {
    if (!agent || nextSession.id === session?.id) {
      return;
    }
    if (session?.id) {
      persistCurrentSnapshot(session.id);
    }
    disconnect();
    setSending(false);
    setError('');
    setNotice('');
    onChatStateChange?.({ activeChatSessionId: nextSession.id, chatHistoryOpen });
    await loadChat(agent.id, nextSession.id, () => true);
    if (options.notice) {
      setNotice(options.notice);
    }
    if (running) {
      connect(agent, nextSession.id);
    }
  }

  function toggleHistory() {
    onChatStateChange?.({ activeChatSessionId: activeSessionId || preferredSessionId, chatHistoryOpen: !chatHistoryOpen });
  }

  function applyEmptySession(nextSession: ChatSessionInfo) {
    renderSequenceRef.current = 0;
    sessionRef.current = nextSession;
    messagesStateRef.current = [];
    interactionsRef.current = [];
    interactionDraftsRef.current = {};
    inputRef.current = '';
    statusRef.current = nextSession.runtimeState || nextSession.status || 'idle';
    errorRef.current = '';
    noticeRef.current = nextSession.lastError || '';
    setSession(nextSession);
    setMessages([]);
    setInteractions([]);
    setInteractionDrafts({});
    setInteractionSubmitting({});
    setInput('');
    setStatus(nextSession.runtimeState || nextSession.status || 'idle');
    setNotice(nextSession.lastError || '');
    setError('');
    chatSnapshots.set(nextSession.id, {
      session: nextSession,
      messages: [],
      interactions: [],
      interactionDrafts: {},
      input: '',
      status: nextSession.runtimeState || nextSession.status || 'idle',
      error: '',
      notice: nextSession.lastError || '',
      renderSequence: 0,
    });
  }

  function touchSessionFromMessage(message: ChatMessageInfo) {
    setSessions((current) => current.map((item) => {
      if (item.id !== message.sessionId) {
        return item;
      }
      const title = item.title === '新对话' && message.role === 'user'
        ? truncateText(message.content, 32) || item.title
        : item.title;
      return {
        ...item,
        title,
        lastActiveAt: Date.now(),
        lastMessagePreview: truncateText(message.content, 80) || item.lastMessagePreview,
        messageCount: Math.max(item.messageCount, message.sequence),
      };
    }).sort((a, b) => b.lastActiveAt - a.lastActiveAt));
  }

  return (
    <div
      className={cn(
        'grid h-full min-h-0 min-w-0 max-w-full overflow-hidden max-[980px]:max-w-[100dvw]',
        chatHistoryOpen ? 'grid-cols-[minmax(0,1fr)_300px] max-[760px]:grid-cols-1' : 'grid-cols-1',
        active ? '' : 'hidden',
      )}
      data-testid="chat-panel"
    >
      <section className="grid h-full min-h-0 min-w-0 max-w-full grid-rows-[auto_minmax(0,1fr)_auto_auto_auto] overflow-hidden max-[980px]:max-w-[100dvw]">
        <div className="border-b border-border px-3 py-2">
          <div className="flex min-w-0 items-center justify-between gap-2">
            <h2 className="truncate text-sm font-semibold text-foreground">{session?.title || '新对话'}</h2>
            <div className="flex shrink-0 items-center gap-1">
              <IconButton disabled={!agent || loading} title="新建对话" onClick={() => void createSession()}>
                <Plus className="h-4 w-4" />
              </IconButton>
              <IconButton aria-pressed={chatHistoryOpen} title="对话历史" onClick={toggleHistory}>
                <History className="h-4 w-4" />
              </IconButton>
            </div>
          </div>
          <div className="mt-1 flex min-h-5 items-center gap-2 text-xs text-muted-foreground">
            <span className={cn('dot', running && status !== 'error' ? (processing || waitingInput ? 'working' : 'online') : 'offline')} />
            <span>{running ? connectionStatusLabel(status, socketReady, processing, waitingInput) : '离线'}</span>
            {session ? <span className="chat-session-id truncate">{session.id}</span> : null}
          </div>
        </div>
        <ChatThread ref={messagesRef}>
          {loading && messages.length === 0 ? <EmptyState icon={<Loader2 className="h-5 w-5" />} title="正在加载对话" /> : null}
          {!loading && !agent ? <EmptyState icon={<Bot className="h-5 w-5" />} title="未选择智能体" /> : null}
          {!loading && agent && orderedMessages.length === 0 ? (
            <EmptyState
              icon={<MessageSquare className="h-5 w-5" />}
              title={running ? '开始与该智能体对话' : '启动智能体后即可对话'}
              description="对话历史和终端输出会在这里按时间线展示。"
            />
          ) : null}
          {orderedMessages.map((message) => (
            <ChatMessage
              key={message.id}
              label={messageLabel(message, agent)}
              message={message}
              statusLabel={message.role === 'terminal' ? '实时输出' : statusLabel(message.status)}
            >
              {message.role === 'assistant'
                ? <MarkdownContent content={message.content || (message.status === 'streaming' ? '...' : '')} />
                : <p className="m-0 whitespace-pre-wrap break-words leading-6">{message.content || (message.status === 'streaming' ? '...' : '')}</p>}
            </ChatMessage>
          ))}
        </ChatThread>
        {pendingInteractions.length ? (
          <div className="grid max-h-[32dvh] gap-2 overflow-auto border-t border-border bg-muted/20 p-3 max-[640px]:max-h-52" data-testid="chat-interactions">
            {pendingInteractions.map((interaction) => (
              <ChatInteractionCard
                active={interaction.id === firstPendingInteractionId}
                draft={interactionDrafts[interaction.id] ?? ''}
                interaction={interaction}
                key={interaction.id}
                submitting={Boolean(interactionSubmitting[interaction.id])}
                onCancel={() => void cancelInteraction(interaction)}
                onDraft={(value) => setInteractionDrafts((current) => ({ ...current, [interaction.id]: value }))}
                onSubmit={(response, mode, scope) => void submitInteraction(interaction, response, mode, scope)}
              />
            ))}
          </div>
        ) : null}
        {error || notice ? (
          <Alert tone={error ? 'danger' : 'info'} className="m-3 w-fit max-w-[calc(100%-1.5rem)]" onDismiss={() => { setError(''); setNotice(''); }}>
            <div className="flex min-w-0 flex-wrap items-center gap-2">
              <span className="min-w-0 break-words" data-testid="chat-alert-message">{error || notice}</span>
              {canRecover ? (
                <Button size="sm" type="button" variant="soft" onClick={() => void recoverChat()}>
                  <RotateCcw className="h-4 w-4" />
                  从历史继续
                </Button>
              ) : null}
            </div>
          </Alert>
        ) : null}
        <ChatComposer
          data-testid="chat-composer"
          onSubmit={(event) => {
            event.preventDefault();
            sendMessage();
          }}
        >
          <Textarea
            aria-label="输入对话内容"
            className="max-h-[30dvh] min-h-16 resize-none overflow-y-auto max-[640px]:max-h-24"
            disabled={sendBlocked}
            placeholder={waitingInput ? '请先处理当前交互请求' : '输入对话内容'}
            value={input}
            onChange={(event) => setInput(event.target.value)}
            onKeyDown={(event) => {
              if (event.key === 'Enter' && !event.shiftKey) {
                event.preventDefault();
                sendMessage();
              }
            }}
          />
          <Button className="h-full max-[640px]:w-full" disabled={!canSend} type="submit" variant="primary">
            <Send className="h-4 w-4" />
            发送
          </Button>
        </ChatComposer>
      </section>
      {chatHistoryOpen ? (
        <ChatHistorySidebar
          activeSessionId={activeSessionId}
          sessions={sessions}
          onClose={toggleHistory}
          onDelete={setDeleteTarget}
          onSelect={(item) => void selectSession(item)}
        />
      ) : null}
      <ConfirmDialog
        danger
        confirmText="删除"
        description={`删除“${deleteTarget?.title || '新对话'}”后，该对话的历史消息也会一起删除。`}
        disabled={loading}
        open={Boolean(deleteTarget)}
        title="删除对话"
        onClose={() => setDeleteTarget(null)}
        onConfirm={() => {
          if (deleteTarget) {
            void deleteSession(deleteTarget);
          }
        }}
      />
    </div>
  );
}

function defaultExecutionEvents(message: ChatMessageInfo): ChatExecutionEvent[] {
  if (message.role !== 'assistant' || message.status !== 'streaming') {
    return [];
  }
  return [{
    id: `local-${message.id}-processing`,
    kind: 'status',
    title: '正在处理',
    status: 'running',
    createdAt: new Date().toISOString(),
  }];
}

function persistSnapshot(agentId: string, snapshot: ChatSnapshot) {
  chatSnapshots.set(agentId, {
    ...snapshot,
    messages: snapshot.messages.map((message) => ({ ...message, executionEvents: [...(message.executionEvents ?? [])] })),
    interactions: snapshot.interactions.map((interaction) => ({ ...interaction })),
    interactionDrafts: { ...snapshot.interactionDrafts },
  });
}

function parseIncomingChatEvent(data: unknown): ChatEvent | null {
  const text = readIncomingText(data);
  if (!text) {
    return null;
  }
  try {
    const value = JSON.parse(text) as unknown;
    if (value && typeof value === 'object' && typeof (value as { type?: unknown }).type === 'string') {
      return value as ChatEvent;
    }
  } catch {
    return null;
  }
  return null;
}

function readIncomingText(data: unknown): string {
  if (typeof data === 'string') {
    return data;
  }
  if (data instanceof Blob) {
    return '';
  }
  if (data instanceof ArrayBuffer) {
    return new TextDecoder().decode(data);
  }
  if (ArrayBuffer.isView(data)) {
    return new TextDecoder().decode(data);
  }
  return String(data ?? '');
}

function showChatDeleteToast() {
  toast.success('已删除对话。', {
    closeButton: true,
    duration: Infinity,
    id: 'chat-session-deleted',
  });
}

function ChatInteractionCard({
  active,
  draft,
  interaction,
  submitting,
  onCancel,
  onDraft,
  onSubmit,
}: {
  active: boolean;
  draft: string;
  interaction: ChatInteractionInfo;
  submitting: boolean;
  onCancel: () => void;
  onDraft: (value: string) => void;
  onSubmit: (response: Record<string, unknown>, mode: string, scope?: ChatInteractionResponseScope) => void;
}) {
  const payload = parseInteractionPayload(interaction.payload);
  const disabled = !active || interaction.status !== 'pending' || submitting;
  const title = interaction.title || interactionTypeLabel(interaction.type);
  return (
    <section
      className={cn(
        'grid gap-3 rounded-[8px] border bg-card p-3 text-sm shadow-xs',
        interaction.status === 'pending' && 'border-primary/30',
        interaction.status !== 'pending' && 'opacity-75',
      )}
      data-interaction-status={interaction.status}
      data-testid="chat-interaction-card"
    >
      <div className="flex min-w-0 flex-wrap items-start justify-between gap-2">
        <div className="flex min-w-0 items-start gap-2">
          <span className="mt-0.5 grid size-7 shrink-0 place-items-center rounded-md bg-primary/10 text-primary">
            {interaction.type === 'permission' ? <ShieldCheck className="size-4" /> : <AlertTriangle className="size-4" />}
          </span>
          <div className="grid min-w-0 gap-1">
            <div className="flex min-w-0 flex-wrap items-center gap-2">
              <strong className="truncate text-sm font-semibold">{title}</strong>
              <Badge tone={riskTone(interaction.riskLevel)}>{riskLabel(interaction.riskLevel)}</Badge>
              <Badge tone={interaction.status === 'pending' ? 'warning' : interaction.status === 'resolved' ? 'success' : 'neutral'}>{interactionStatusLabel(interaction.status)}</Badge>
            </div>
            <p className="m-0 min-w-0 whitespace-pre-wrap break-words text-xs leading-5 text-muted-foreground">{interaction.body || payloadSummary(payload)}</p>
          </div>
        </div>
      </div>
      {interaction.status === 'pending' ? (
        <InteractionControls
          disabled={disabled}
          draft={draft}
          interaction={interaction}
          payload={payload}
          submitting={submitting}
          onCancel={onCancel}
          onDraft={onDraft}
          onSubmit={onSubmit}
        />
      ) : interaction.response ? (
        <div className="rounded-md bg-muted px-2 py-1.5 text-xs text-muted-foreground">{interactionResponseLabel(interaction)}</div>
      ) : null}
    </section>
  );
}

function InteractionControls({
  disabled,
  draft,
  interaction,
  payload,
  submitting,
  onCancel,
  onDraft,
  onSubmit,
}: {
  disabled: boolean;
  draft: string;
  interaction: ChatInteractionInfo;
  payload: Record<string, unknown>;
  submitting: boolean;
  onCancel: () => void;
  onDraft: (value: string) => void;
  onSubmit: (response: Record<string, unknown>, mode: string, scope?: ChatInteractionResponseScope) => void;
}) {
  if (interaction.type === 'permission') {
    return (
      <div className="flex min-w-0 flex-wrap gap-2">
        <Button disabled={disabled} size="sm" type="button" variant="primary" onClick={() => onSubmit({ decision: 'allow' }, 'allow', 'once')}>
          <Check className="h-4 w-4" />
          允许一次
        </Button>
        <Button disabled={disabled} size="sm" type="button" variant="soft" onClick={() => onSubmit({ decision: 'allow' }, 'allow_session', 'session')}>
          允许本会话
        </Button>
        <Button disabled={disabled} size="sm" type="button" variant="danger" onClick={() => onSubmit({ decision: 'reject' }, 'reject', 'once')}>
          拒绝
        </Button>
        <Button disabled={submitting} size="sm" type="button" variant="soft" onClick={onCancel}>
          取消
        </Button>
      </div>
    );
  }
  if (interaction.type === 'choice') {
    const question = firstInteractionQuestion(payload);
    const selected = selectedChoiceValues(draft, question);
    if (question.options.length === 0) {
      return (
        <div className="grid gap-2">
          <Input
            aria-label="交互响应"
            disabled={disabled}
            placeholder="输入响应"
            value={draft}
            onChange={(event) => onDraft(event.target.value)}
          />
          <InteractionFooter disabled={disabled || !draft.trim()} submitting={submitting} onCancel={onCancel} onSubmit={() => onSubmit({ answer: draft.trim() }, 'answer', 'once')} />
        </div>
      );
    }
    return (
      <div className="grid gap-2">
        <div className="grid gap-1">
          {question.options.map((option) => (
            <CheckboxField
              checked={selected.includes(option.label)}
              disabled={disabled}
              key={option.label}
              onChange={() => {
                if (question.multiSelect) {
                  onDraft(toggleChoiceValue(selected, option.label).join('\n'));
                  return;
                }
                onDraft(option.label);
              }}
            >
              <span className="grid gap-0.5">
                <span>{option.label}</span>
                {option.description ? <span className="text-xs text-muted-foreground">{option.description}</span> : null}
              </span>
            </CheckboxField>
          ))}
        </div>
        <InteractionFooter
          disabled={disabled || selected.length === 0}
          submitting={submitting}
          onCancel={onCancel}
          onSubmit={() => onSubmit({ answer: question.multiSelect ? selected : selected[0] }, 'answer', 'once')}
        />
      </div>
    );
  }
  if (interaction.type === 'auth' || interaction.type === 'plan') {
    const suggested = interaction.type === 'plan' ? 'accepted' : 'completed';
    return (
      <div className="grid gap-2">
        <Input
          aria-label="交互响应"
          disabled={disabled}
          placeholder={interaction.type === 'plan' ? '可填写计划调整说明' : '认证完成后输入说明'}
          value={draft}
          onChange={(event) => onDraft(event.target.value)}
        />
        <div className="flex min-w-0 flex-wrap gap-2">
          <Button disabled={disabled} size="sm" type="button" variant="primary" onClick={() => onSubmit({ answer: draft.trim(), decision: suggested }, 'answer', 'once')}>
            {submitting ? <Loader2 className="h-4 w-4 animate-spin" /> : <Check className="h-4 w-4" />}
            {interaction.type === 'plan' ? '接受' : '已完成'}
          </Button>
          <Button disabled={disabled} size="sm" type="button" variant="danger" onClick={() => onSubmit({ decision: 'reject', answer: draft.trim() }, 'reject', 'once')}>
            拒绝
          </Button>
          <Button disabled={submitting} size="sm" type="button" variant="soft" onClick={onCancel}>
            取消
          </Button>
        </div>
      </div>
    );
  }
  return (
    <div className="grid gap-2">
      <Input
        aria-label="交互响应"
        disabled={disabled}
        placeholder="输入响应"
        value={draft}
        onChange={(event) => onDraft(event.target.value)}
      />
      <InteractionFooter disabled={disabled || !draft.trim()} submitting={submitting} onCancel={onCancel} onSubmit={() => onSubmit({ answer: draft.trim() }, 'answer', 'once')} />
    </div>
  );
}

function InteractionFooter({ disabled, submitting, onCancel, onSubmit }: { disabled: boolean; submitting: boolean; onCancel: () => void; onSubmit: () => void }) {
  return (
    <div className="flex min-w-0 flex-wrap gap-2">
      <Button disabled={disabled} size="sm" type="button" variant="primary" onClick={onSubmit}>
        {submitting ? <Loader2 className="h-4 w-4 animate-spin" /> : <Check className="h-4 w-4" />}
        提交
      </Button>
      <Button disabled={submitting} size="sm" type="button" variant="soft" onClick={onCancel}>
        取消
      </Button>
    </div>
  );
}

function ChatHistorySidebar({
  activeSessionId,
  sessions,
  onClose,
  onDelete,
  onSelect,
}: {
  activeSessionId: string;
  sessions: ChatSessionInfo[];
  onClose: () => void;
  onDelete: (session: ChatSessionInfo) => void;
  onSelect: (session: ChatSessionInfo) => void;
}) {
  return (
    <aside className="grid h-full min-h-0 grid-rows-[auto_minmax(0,1fr)] border-l border-border bg-muted/30 max-[760px]:border-l-0 max-[760px]:border-t" data-testid="chat-history-sidebar">
      <div className="flex min-h-11 items-center justify-between gap-2 border-b border-border px-3">
        <h3 className="text-sm font-semibold">对话历史</h3>
        <IconButton title="关闭对话历史" onClick={onClose}>
          <X className="h-4 w-4" />
        </IconButton>
      </div>
      <div className="min-h-0 overflow-auto p-2">
        {sessions.length === 0 ? (
          <div className="rounded-[6px] border border-dashed border-border p-4 text-center text-sm text-muted-foreground">暂无对话</div>
        ) : null}
        <div className="flex flex-col gap-2">
          {sessions.map((item) => {
            const waiting = isWaitingSession(item);
            const dotState = chatSessionDotState(item);
            const deletingDisabled = isProcessingSession(item) || waiting;
            return (
            <div
              className={cn(
                'grid min-h-[76px] w-full grid-cols-[minmax(0,1fr)_auto] items-start gap-2 rounded-[6px] border border-border bg-card p-2 text-left text-sm hover:bg-accent',
                item.id === activeSessionId && 'border-primary/30 bg-primary/10 hover:bg-primary/15',
              )}
              data-testid="chat-history-row"
              key={item.id}
            >
              <Button
                className="grid h-auto min-w-0 grid-cols-[10px_minmax(0,1fr)] items-start justify-start gap-2 bg-transparent p-0 text-left font-normal hover:bg-transparent"
                data-testid="chat-history-item"
                type="button"
                variant="ghost"
                onClick={() => onSelect(item)}
              >
                <span
                  className={cn('dot mt-1.5', dotState)}
                  data-chat-session-state={dotState}
                  data-testid="chat-history-status-dot"
                />
                <span className="grid min-w-0 gap-1">
                  <strong className="truncate text-sm font-medium">{item.title || '新对话'}</strong>
                  <span className="truncate text-xs leading-5 text-muted-foreground">{item.lastMessagePreview || '尚无消息'}</span>
                  <span className="truncate text-[11px] text-muted-foreground">{formatSessionTime(item.lastActiveAt)} · {sessionStateLabel(item)}</span>
                </span>
              </Button>
              <IconButton
                data-testid="chat-history-delete"
                disabled={deletingDisabled}
                title={deletingDisabled ? '工作中或待确认对话不可删除' : '删除对话'}
                onClick={() => onDelete(item)}
              >
                <Trash2 />
              </IconButton>
            </div>
          );})}
        </div>
      </div>
    </aside>
  );
}

function findReusableEmptySession(items: ChatSessionInfo[], currentSessionId?: string) {
  return items.find((item) => item.messageCount === 0 && item.id !== currentSessionId)
    ?? items.find((item) => item.messageCount === 0)
    ?? null;
}

function upsertChatSession(items: ChatSessionInfo[], session: ChatSessionInfo) {
  const next = [session, ...items.filter((item) => item.id !== session.id)];
  return next.sort((a, b) => b.lastActiveAt - a.lastActiveAt);
}

function truncateText(value: string, maxLength: number) {
  const normalized = value.trim().replace(/\s+/g, ' ');
  if (normalized.length <= maxLength) {
    return normalized;
  }
  return normalized.slice(0, maxLength);
}

function formatSessionTime(value: number) {
  if (!value) {
    return '刚刚';
  }
  return new Intl.DateTimeFormat(undefined, {
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
  }).format(new Date(value));
}

function sessionStateLabel(session: ChatSessionInfo) {
  if (isWaitingSession(session)) {
    return '待确认';
  }
  if (isProcessingSession(session)) {
    return '工作中';
  }
  if (session.runtimeState === 'error' || session.status === 'error') {
    return '异常';
  }
  if (session.runtimeState === 'exited' || session.status === 'exited') {
    return '已退出';
  }
  return '空闲';
}

function chatSessionDotState(session: ChatSessionInfo) {
  if (isWaitingSession(session)) {
    return 'waiting';
  }
  if (isProcessingSession(session)) {
    return 'working';
  }
  if (session.runtimeState === 'error' || session.status === 'error') {
    return 'error';
  }
  if (session.runtimeState === 'exited' || session.status === 'exited') {
    return 'exited';
  }
  return 'idle';
}

function isProcessingSession(session: ChatSessionInfo) {
  return session.runtimeState === 'running' || session.runtimeState === 'recovering' || session.status === 'running' || session.status === 'recovering';
}

function isWaitingSession(session: ChatSessionInfo) {
  return session.runtimeState === 'waiting_input' || session.status === 'waiting_input';
}

function isProcessingStatus(value: string) {
  return value === 'running' || value === 'recovering' || value === 'waiting_input';
}

function isInteractionProgressMessage(message: RenderedChatMessage) {
  return Boolean(message.transient && message.terminalId === interactionProgressTerminalId);
}

function isEmptyStreamingAssistantPlaceholder(message: RenderedChatMessage) {
  return message.role === 'assistant'
    && message.status === 'streaming'
    && message.content.trim() === ''
    && (message.executionEvents ?? []).every(isLocalProcessingEvent);
}

function mergeInteractions(current: ChatInteractionInfo[], incoming: ChatInteractionInfo[]) {
  const byId = new Map<string, ChatInteractionInfo>();
  for (const item of current) {
    byId.set(item.id, item);
  }
  for (const item of incoming) {
    byId.set(item.id, item);
  }
  return [...byId.values()].sort((a, b) => a.createdAt - b.createdAt);
}

type InteractionOption = {
  label: string;
  description?: string;
  preview?: string;
};

type InteractionQuestion = {
  question: string;
  header?: string;
  options: InteractionOption[];
  multiSelect: boolean;
};

function parseInteractionPayload(raw: string): Record<string, unknown> {
  if (!raw) {
    return {};
  }
  try {
    const value = JSON.parse(raw) as unknown;
    if (value && typeof value === 'object' && !Array.isArray(value)) {
      return value as Record<string, unknown>;
    }
  } catch {
    return {};
  }
  return {};
}

function firstInteractionQuestion(payload: Record<string, unknown>): InteractionQuestion {
  const questions = Array.isArray(payload.questions) ? payload.questions : [];
  const rawQuestion = firstObject(questions) ?? payload;
  const question = typeof rawQuestion.question === 'string'
    ? rawQuestion.question
    : typeof rawQuestion.prompt === 'string'
    ? rawQuestion.prompt
    : '';
  const options: InteractionOption[] = [];
  for (const item of Array.isArray(rawQuestion.options) ? rawQuestion.options : []) {
    if (!item || typeof item !== 'object') {
      continue;
    }
    const option = item as Record<string, unknown>;
    const label = typeof option.label === 'string'
      ? option.label
      : typeof option.value === 'string'
      ? option.value
      : '';
    if (!label) {
      continue;
    }
    options.push({
      label,
      description: typeof option.description === 'string' ? option.description : undefined,
      preview: typeof option.preview === 'string' ? option.preview : undefined,
    });
  }
  return {
    question,
    header: typeof rawQuestion.header === 'string' ? rawQuestion.header : undefined,
    options,
    multiSelect: Boolean(rawQuestion.multiSelect),
  };
}

function firstObject(items: unknown[]): Record<string, unknown> | null {
  for (const item of items) {
    if (item && typeof item === 'object' && !Array.isArray(item)) {
      return item as Record<string, unknown>;
    }
  }
  return null;
}

function selectedChoiceValues(draft: string, question: InteractionQuestion) {
  const selected = draft
    .split('\n')
    .map((item) => item.trim())
    .filter(Boolean);
  if (selected.length > 0) {
    return selected;
  }
  const first = firstOptionLabel(question);
  return first ? [first] : [];
}

function firstOptionLabel(question: InteractionQuestion) {
  return question.options[0]?.label ?? '';
}

function toggleChoiceValue(current: string[], value: string) {
  return current.includes(value)
    ? current.filter((item) => item !== value)
    : [...current, value];
}

function payloadSummary(payload: Record<string, unknown>) {
  const toolName = textValue(payload, 'toolName', 'tool_name', 'name');
  const target = textValue(payload, 'target', 'command', 'path', 'pattern', 'summary');
  const input = textValue(payload, 'inputSummary', 'input_summary', 'description', 'preview');
  return [toolName, target, input].filter(Boolean).join(' · ') || '底层工具正在等待用户输入。';
}

function textValue(payload: Record<string, unknown>, ...keys: string[]) {
  for (const key of keys) {
    const value = payload[key];
    if (typeof value === 'string' && value.trim()) {
      return value.trim();
    }
  }
  return '';
}

function interactionTypeLabel(value: string) {
  const labels: Record<string, string> = {
    permission: '权限确认',
    question: '需要确认',
    choice: '请选择',
    text: '需要输入',
    auth: '认证介入',
    plan: '计划确认',
    custom: '交互请求',
  };
  return labels[value] ?? '交互请求';
}

function riskTone(value: string): 'neutral' | 'success' | 'warning' | 'danger' | 'info' | 'primary' {
  if (value === 'critical' || value === 'high') {
    return 'danger';
  }
  if (value === 'medium') {
    return 'warning';
  }
  if (value === 'low') {
    return 'info';
  }
  return 'neutral';
}

function riskLabel(value: string) {
  const labels: Record<string, string> = {
    low: '低风险',
    medium: '中风险',
    high: '高风险',
    critical: '关键风险',
  };
  return labels[value] ?? '未知风险';
}

function interactionStatusLabel(value: string) {
  const labels: Record<string, string> = {
    pending: '待处理',
    resolved: '已完成',
    rejected: '已拒绝',
    cancelled: '已取消',
    expired: '已过期',
    error: '异常',
  };
  return labels[value] ?? value;
}

function interactionResponseLabel(interaction: ChatInteractionInfo) {
  if (interaction.status === 'rejected') {
    return '已拒绝该请求';
  }
  if (interaction.status === 'cancelled') {
    return '已取消该请求';
  }
  if (interaction.status === 'expired') {
    return '该请求已过期';
  }
  if (interaction.status === 'error') {
    return '该请求无法继续';
  }
  const response = parseInteractionPayload(interaction.response);
  const decision = textValue(response, 'decision', 'answer', 'value');
  if (decision) {
    return `已提交：${decision}`;
  }
  return '已提交响应';
}

function mergeLoadedMessages(loadedMessages: RenderedChatMessage[], cachedMessages: RenderedChatMessage[]) {
  const nextById = new Map<number, RenderedChatMessage>();

  for (const loadedMessage of loadedMessages) {
    nextById.set(loadedMessage.id, loadedMessage);
  }

  for (const cachedMessage of cachedMessages) {
    const loadedMessage = nextById.get(cachedMessage.id);
    if (!loadedMessage) {
      nextById.set(cachedMessage.id, cachedMessage);
      continue;
    }
    nextById.set(cachedMessage.id, {
      ...loadedMessage,
      content: cachedMessage.content.length > loadedMessage.content.length ? cachedMessage.content : loadedMessage.content,
      status: cachedMessage.status === 'streaming' && loadedMessage.status === 'complete' ? loadedMessage.status : cachedMessage.status,
      renderSequence: cachedMessage.renderSequence,
      executionEvents: cachedMessage.executionEvents ?? loadedMessage.executionEvents ?? [],
      terminalId: cachedMessage.terminalId,
      transient: cachedMessage.transient,
    });
  }

  return [...nextById.values()].sort((a, b) => a.renderSequence - b.renderSequence);
}

function normalizeExecutionEvent(event: ChatExecutionEvent): ChatExecutionEvent | null {
  const title = event.title.trim();
  const detail = (event.detail ?? '').trim();
  if (!title && !detail) {
    return null;
  }
  return {
    ...event,
    title: title || event.kind,
    detail: detail || undefined,
  };
}

function mergeExecutionEvent(existing: ChatExecutionEvent[], event: ChatExecutionEvent) {
  const baseEvents = existing.filter((item) => !isLocalProcessingEvent(item));
  const idIndex = baseEvents.findIndex((item) => item.id === event.id);
  if (idIndex !== -1) {
    const next = [...baseEvents];
    next[idIndex] = mergeExecutionEventById(next[idIndex], event);
    return next;
  }

  if (isCollapsibleExecutionEvent(event)) {
    const fingerprint = executionEventFingerprint(event);
    const duplicateIndex = baseEvents.findIndex((item) => (
      isCollapsibleExecutionEvent(item) && executionEventFingerprint(item) === fingerprint
    ));
    if (duplicateIndex !== -1) {
      const next = [...baseEvents];
      next[duplicateIndex] = event;
      return next;
    }
  }

  return [...baseEvents, event];
}

function mergeExecutionEventById(existing: ChatExecutionEvent, event: ChatExecutionEvent): ChatExecutionEvent {
  if (event.kind === 'thinking' && !event.detail) {
    return existing;
  }
  return event;
}

function isLocalProcessingEvent(event: ChatExecutionEvent) {
  return event.id.startsWith('local-') && event.kind === 'status' && event.title === '正在处理';
}

function isCollapsibleExecutionEvent(event: ChatExecutionEvent) {
  return event.kind === 'status' || event.kind === 'thinking';
}

function executionEventFingerprint(event: ChatExecutionEvent) {
  return [
    event.kind,
    event.title.trim(),
    (event.detail ?? '').trim(),
  ].join('\u0001');
}

function MarkdownContent({ content }: { content: string }) {
  return <div className="min-w-0 space-y-1 text-sm leading-6">{parseMarkdownBlocks(content)}</div>;
}

function parseMarkdownBlocks(text: string): ReactNode[] {
  const lines = text.split('\n');
  const blocks: ReactNode[] = [];
  let i = 0;

  while (i < lines.length) {
    const line = lines[i];

    // Fenced code block
    if (line.startsWith('```')) {
      const closingFenceIndex = findClosingCodeFence(lines, i);
      if (closingFenceIndex === -1) {
        blocks.push(<p key={blocks.length}>{parseInline(line)}</p>);
        i++;
        continue;
      }
      const codeLines = lines.slice(i + 1, closingFenceIndex);
      blocks.push(
        <div key={blocks.length} className="overflow-hidden rounded-[8px] border border-border">
          <div className="flex items-center gap-2 border-b border-border bg-muted px-3 py-1.5 text-[11px] text-muted-foreground">
            <span className="font-mono">code</span>
          </div>
          <pre className="overflow-x-auto bg-muted px-3 py-2 font-mono text-xs leading-5">
            <code>{codeLines.join('\n')}</code>
          </pre>
        </div>,
      );
      i = closingFenceIndex + 1;
      continue;
    }

    // Headings
    const headingMatch = line.match(/^(#{1,3})\s+(.*)$/);
    if (headingMatch) {
      blocks.push(
        <p key={blocks.length} className="font-semibold leading-snug text-foreground">
          {parseInline(headingMatch[2])}
        </p>,
      );
      i++;
      continue;
    }

    // Bullet list
    if (/^[-*+] /.test(line)) {
      const items: ReactNode[] = [];
      while (i < lines.length && /^[-*+] /.test(lines[i])) {
        items.push(<li key={i}>{parseInline(lines[i].replace(/^[-*+] /, ''))}</li>);
        i++;
      }
      blocks.push(<ul key={blocks.length} className="list-disc space-y-0.5 pl-4">{items}</ul>);
      continue;
    }

    // Numbered list
    if (/^\d+[.)] /.test(line)) {
      const items: ReactNode[] = [];
      while (i < lines.length && /^\d+[.)] /.test(lines[i])) {
        items.push(<li key={i}>{parseInline(lines[i].replace(/^\d+[.)] /, ''))}</li>);
        i++;
      }
      blocks.push(<ol key={blocks.length} className="list-decimal space-y-0.5 pl-4">{items}</ol>);
      continue;
    }

    // Horizontal rule
    if (/^-{3,}$/.test(line.trim())) {
      blocks.push(<hr key={blocks.length} className="border-border" />);
      i++;
      continue;
    }

    // Empty line
    if (line.trim() === '') {
      i++;
      continue;
    }

    // Regular paragraph
    blocks.push(<p key={blocks.length}>{parseInline(line)}</p>);
    i++;
  }

  return blocks;
}

function findClosingCodeFence(lines: string[], openingIndex: number) {
  for (let index = openingIndex + 1; index < lines.length; index += 1) {
    if (lines[index].startsWith('```')) {
      return index;
    }
  }
  return -1;
}

function parseInline(text: string): ReactNode {
  const TOKEN = /(\*\*(?:[^*]|\*(?!\*))+\*\*|\*(?:[^*\n])+\*|`[^`\n]+`)/g;
  const parts = text.split(TOKEN);
  if (parts.length === 1) return text;
  return (
    <>
      {parts.map((part, j) => {
        if (!part) return null;
        if (part.startsWith('**') && part.endsWith('**') && part.length > 4)
          return <strong key={j}>{part.slice(2, -2)}</strong>;
        if (part.startsWith('`') && part.endsWith('`') && part.length > 2)
          return <code key={j} className="rounded bg-muted px-1 font-mono text-xs">{part.slice(1, -1)}</code>;
        if (part.startsWith('*') && part.endsWith('*') && part.length > 2)
          return <em key={j}>{part.slice(1, -1)}</em>;
        return part;
      })}
    </>
  );
}

function messageLabel(message: ChatMessageInfo, agent?: AgentInfo) {
  if (message.role === 'user') return '你';
  if (message.role === 'assistant') return agent?.name || '智能体';
  if (message.role === 'terminal') return '终端';
  if (message.role === 'error') return '错误';
  return '系统';
}

function statusLabel(value: string) {
  const labels: Record<string, string> = {
    idle: '空闲',
    running: '已连接',
    waiting_input: '待确认',
    connected: '已连接',
    connecting: '连接中',
    closed: '已关闭',
    offline: '离线',
    error: '异常',
    exited: '已退出',
    complete: '完成',
    streaming: '生成中',
  };
  return labels[value] ?? value;
}

function connectionStatusLabel(value: string, socketReady: boolean, processing: boolean, waitingInput: boolean) {
  if (waitingInput || value === 'waiting_input') {
    return '待确认';
  }
  if (processing) {
    return '工作中';
  }
  if (socketReady && (value === 'idle' || value === 'connected' || value === 'running')) {
    return '已连接';
  }
  return statusLabel(value);
}
