/**
 * Event Bus Store - Manages application-wide events using Svelte 5 runes
 *
 * Event types:
 * - progress: File upload/download progress
 * - notification: Success/error/info messages
 * - sync: P2P synchronization status (future)
 */

import { EventsOn, EventsOff } from '../wails/wailsjs/runtime/runtime';

export type NotificationType = 'success' | 'error' | 'info' | 'warning';

export interface Notification {
  id: string;
  type: NotificationType;
  title: string;
  message?: string;
  timestamp: number;
  duration?: number; // Auto-dismiss after X milliseconds (0 = manual dismiss)
}

export interface ProgressEvent {
  id: string;
  operation: string;
  current: number;
  total: number;
  filename?: string;
}

class EventBus {
  notifications = $state<Notification[]>([]);
  progressEvents = $state<Map<string, ProgressEvent>>(new Map());

  constructor() {
    // Listen to backend events
    this.initWailsListeners();
  }

  private initWailsListeners() {
    // Listen to backend notifications
    EventsOn('notification', (data: any) => {
      this.addNotification({
        type: data.type || 'info',
        title: data.title,
        message: data.message,
        duration: data.duration || 5000,
      });
    });

    // Listen to backend progress events
    EventsOn('progress', (data: any) => {
      this.updateProgress(data.id, data);
    });

    // Listen to operation completion
    EventsOn('progress:complete', (data: any) => {
      this.removeProgress(data.id);
    });
  }

  // Notifications management
  addNotification(params: Omit<Notification, 'id' | 'timestamp'>) {
    const notification: Notification = {
      id: crypto.randomUUID(),
      timestamp: Date.now(),
      ...params,
    };

    this.notifications = [...this.notifications, notification];

    // Auto-dismiss if duration is set
    if (notification.duration && notification.duration > 0) {
      setTimeout(() => {
        this.removeNotification(notification.id);
      }, notification.duration);
    }

    return notification.id;
  }

  removeNotification(id: string) {
    this.notifications = this.notifications.filter(n => n.id !== id);
  }

  clearNotifications() {
    this.notifications = [];
  }

  // Progress tracking
  updateProgress(id: string, event: Omit<ProgressEvent, 'id'>) {
    const updated = new Map(this.progressEvents);
    updated.set(id, { id, ...event });
    this.progressEvents = updated;
  }

  removeProgress(id: string) {
    const updated = new Map(this.progressEvents);
    updated.delete(id);
    this.progressEvents = updated;
  }

  clearProgress() {
    this.progressEvents = new Map();
  }

  // Convenience methods
  success(title: string, message?: string, duration = 5000) {
    return this.addNotification({ type: 'success', title, message, duration });
  }

  error(title: string, message?: string, duration = 8000) {
    return this.addNotification({ type: 'error', title, message, duration });
  }

  info(title: string, message?: string, duration = 5000) {
    return this.addNotification({ type: 'info', title, message, duration });
  }

  warning(title: string, message?: string, duration = 6000) {
    return this.addNotification({ type: 'warning', title, message, duration });
  }

  // Cleanup
  destroy() {
    EventsOff('notification');
    EventsOff('progress');
    EventsOff('progress:complete');
  }
}

// Singleton instance
export const eventBus = new EventBus();
