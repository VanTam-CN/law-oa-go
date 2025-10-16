import React, { useState, useEffect } from 'react';
import { Spinner } from 'react-bootstrap';
import { getUserSettings, updateUserSettings } from '../services/settingsService';
import { UserSettings } from '../types';

const SettingsPage: React.FC = () => {
  const [settings, setSettings] = useState<UserSettings | null>(null);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [activeSection, setActiveSection] = useState('general');
  const [formData, setFormData] = useState({
    general: {
      name: '',
      email: '',
      language: 'en',
      timezone: 'UTC'
    },
    notifications: {
      email: true,
      sms: false,
      push: true,
      caseUpdates: true,
      clientUpdates: true,
      deadlineReminders: true,
      systemAlerts: true
    },
    privacy: {
      profileVisibility: 'public',
      activityVisibility: 'friends',
      twoFactorAuth: false,
      sessionTimeout: 30
    },
    appearance: {
      theme: 'light',
      fontSize: 'medium',
      dateFormat: 'mm/dd/yyyy',
      timeFormat: '12h'
    }
  });

  useEffect(() => {
    loadSettings();
  }, []);

  const loadSettings = async () => {
    setLoading(true);
    try {
      const response = await getUserSettings();
      // 将settingsService的UserSettings映射到types/index.ts的UserSettings接口
      const mappedSettings: UserSettings = {
        id: response.id,
        user_id: response.user_id,
        language: response.language as 'zh-CN' | 'en-US' | 'zh-TW',
        theme: response.theme,
        timezone: response.timezone,
        date_format: response.date_format as 'YYYY-MM-DD' | 'MM/DD/YYYY' | 'DD/MM/YYYY',
        time_format: response.time_format as '24h' | '12h',
        notifications: {
          email: response.notifications.email,
          push: response.notifications.push,
          sms: response.notifications.sms,
          case_updates: true, // 默认值
          client_updates: true, // 默认值
          system_announcements: true, // 默认值
        },
        privacy: {
          profile_visibility: 'public' as 'public' | 'private' | 'contacts', // 默认值
          activity_tracking: true, // 默认值
          data_sharing: false, // 默认值
        },
        preferences: {
          items_per_page: 10, // 默认值
          default_view: 'list' as 'list' | 'grid' | 'table', // 默认值
          auto_save: true, // 默认值
          confirm_actions: true, // 默认值
          keyboard_shortcuts: true, // 默认值
        },
        created_at: response.created_at,
        updated_at: response.updated_at
      };

      setSettings(mappedSettings);
      setFormData({
        general: {
          name: 'User', // 默认值，因为UserSettings没有name字段
          email: 'user@example.com', // 默认值，因为UserSettings没有email字段
          language: response.language,
          timezone: response.timezone
        },
        notifications: {
          email: response.notifications.email,
          sms: response.notifications.sms,
          push: response.notifications.push,
          caseUpdates: true, // 默认值
          clientUpdates: true, // 默认值
          deadlineReminders: true, // 默认值
          systemAlerts: true // 默认值
        },
        privacy: {
          profileVisibility: response.privacy_settings?.profile_visible ? 'public' : 'private',
          activityVisibility: response.privacy_settings?.activity_tracking ? 'visible' : 'hidden',
          twoFactorAuth: false, // 默认值
          sessionTimeout: 30 // 默认值
        },
        appearance: {
          theme: response.theme,
          fontSize: 'medium', // 默认值
          dateFormat: response.date_format,
          timeFormat: response.time_format
        }
      });
    } catch (error) {
      console.error('Failed to load settings', error);
    } finally {
      setLoading(false);
    }
  };

  const handleChange = (section: string, field: string, value: any) => {
    setFormData(prev => ({
      ...prev,
      [section]: {
        ...prev[section as keyof typeof prev],
        [field]: value
      }
    }));
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setSaving(true);
    try {
      const data: UserSettings = {
        id: settings?.id || 0,
        user_id: settings?.user_id || 0,
        language: formData.general.language as 'zh-CN' | 'en-US' | 'zh-TW',
        theme: formData.appearance.theme as 'light' | 'dark' | 'auto',
        timezone: formData.general.timezone,
        date_format: formData.appearance.dateFormat as 'YYYY-MM-DD' | 'MM/DD/YYYY' | 'DD/MM/YYYY',
        time_format: formData.appearance.timeFormat as '24h' | '12h',
        notifications: {
          email: formData.notifications.email,
          push: formData.notifications.push,
          sms: formData.notifications.sms,
          case_updates: formData.notifications.caseUpdates,
          client_updates: formData.notifications.clientUpdates,
          system_announcements: formData.notifications.systemAlerts,
        },
        privacy: {
          profile_visibility: formData.privacy.profileVisibility as 'public' | 'private' | 'contacts',
          activity_tracking: formData.privacy.activityVisibility === 'visible',
          data_sharing: false,
        },
        preferences: {
          items_per_page: 10,
          default_view: 'list',
          auto_save: true,
          confirm_actions: true,
          keyboard_shortcuts: true
        },
        created_at: settings?.created_at || new Date().toISOString(),
        updated_at: new Date().toISOString()
      };
      const response = await updateUserSettings(data);
      // 将settingsService的UserSettings映射到types/index.ts的UserSettings接口
      const updatedSettings: UserSettings = {
        id: response.id,
        user_id: response.user_id,
        language: response.language as 'zh-CN' | 'en-US' | 'zh-TW',
        theme: response.theme,
        timezone: response.timezone,
        date_format: response.date_format as 'YYYY-MM-DD' | 'MM/DD/YYYY' | 'DD/MM/YYYY',
        time_format: response.time_format as '24h' | '12h',
        notifications: {
          email: response.notifications.email,
          push: response.notifications.push,
          sms: response.notifications.sms,
          case_updates: true,
          client_updates: true,
          system_announcements: true,
        },
        privacy: {
          profile_visibility: 'public',
          activity_tracking: true,
          data_sharing: false,
        },
        preferences: {
          items_per_page: 10,
          default_view: 'list',
          auto_save: true,
          confirm_actions: true,
          keyboard_shortcuts: true,
        },
        created_at: response.created_at,
        updated_at: response.updated_at
      };
      setSettings(updatedSettings);
    } catch (error) {
      console.error('Failed to save settings', error);
    } finally {
      setSaving(false);
    }
  };

  const handleReset = () => {
    if (settings) {
      setFormData({
        general: {
          name: '', // UserSettings接口中没有name字段，使用空字符串作为默认值
          email: '', // UserSettings接口中没有email字段，使用空字符串作为默认值
          language: settings.language,
          timezone: settings.timezone
        },
        notifications: {
          email: settings.notifications.email,
          sms: settings.notifications.sms,
          push: settings.notifications.push,
          caseUpdates: settings.notifications.case_updates, // 映射到正确的字段名
          clientUpdates: settings.notifications.client_updates, // 映射到正确的字段名
          deadlineReminders: true, // UserSettings接口中没有这个字段，使用默认值
          systemAlerts: settings.notifications.system_announcements // 映射到正确的字段名
        },
        privacy: {
          profileVisibility: settings.privacy.profile_visibility, // 映射到正确的字段名
          activityVisibility: 'public', // UserSettings接口中没有这个字段，使用默认值
          twoFactorAuth: false, // UserSettings接口中没有这个字段，使用默认值
          sessionTimeout: 30 // UserSettings接口中没有这个字段，使用默认值
        },
        appearance: {
          theme: settings.theme,
          fontSize: 'medium', // UserSettings接口中没有这个字段，使用默认值
          dateFormat: settings.date_format, // 映射到正确的字段名
          timeFormat: settings.time_format // 映射到正确的字段名
        }
      });
    }
  };

  // Get language display text
  const getLanguageText = (lang: string) => {
    switch (lang) {
      case 'en': return 'English';
      case 'zh': return '中文';
      case 'es': return 'Español';
      case 'fr': return 'Français';
      default: return lang;
    }
  };

  // Get theme display text
  const getThemeText = (theme: string) => {
    switch (theme) {
      case 'light': return 'Light';
      case 'dark': return 'Dark';
      case 'auto': return 'Auto';
      default: return theme;
    }
  };

  // Get font size display text
  const getFontSizeText = (size: string) => {
    switch (size) {
      case 'small': return 'Small';
      case 'medium': return 'Medium';
      case 'large': return 'Large';
      default: return size;
    }
  };

  // Get date format display text
  const getDateFormatText = (format: string) => {
    switch (format) {
      case 'mm/dd/yyyy': return 'MM/DD/YYYY';
      case 'dd/mm/yyyy': return 'DD/MM/YYYY';
      case 'yyyy-mm-dd': return 'YYYY-MM-DD';
      default: return format;
    }
  };

  // Get time format display text
  const getTimeFormatText = (format: string) => {
    switch (format) {
      case '12h': return '12 Hour';
      case '24h': return '24 Hour';
      default: return format;
    }
  };

  // Get visibility display text
  const getVisibilityText = (visibility: string) => {
    switch (visibility) {
      case 'public': return 'Public';
      case 'friends': return 'Friends';
      case 'private': return 'Private';
      default: return visibility;
    }
  };

  if (loading) {
    return (
      <div className="d-flex justify-content-center align-items-center" style={{ height: '50vh' }}>
        <Spinner animation="border" />
        <span className="ms-2">Loading settings...</span>
      </div>
    );
  }

  return (
    <div>
      <div className="d-flex justify-content-between align-items-center mb-4">
        <h1>Settings</h1>
        <div className="d-flex">
          <Button variant="outline-secondary" className="me-2" onClick={handleReset}>
            <i className="fas fa-undo me-2"></i>
            Reset
          </Button>
          <Button variant="primary" onClick={handleSubmit} disabled={saving}>
            {saving ? (
              <span>
                <i className="fas fa-spinner fa-spin me-2"></i>
                Saving...
              </span>
            ) : (
              <span>
                <i className="fas fa-save me-2"></i>
                Save Settings
              </span>
            )}
          </Button>
        </div>
      </div>

      <Row>
        <Col md={3}>
          <Card>
            <Card.Header>
              <i className="fas fa-cog me-2"></i>
              Settings Menu
            </Card.Header>
            <Card.Body>
              <div className="settings-menu">
                <Button
                  variant={activeSection === 'general' ? 'primary' : 'outline-secondary'}
                  className="w-100 mb-2 text-start"
                  onClick={() => setActiveSection('general')}
                >
                  <i className="fas fa-user me-2"></i>
                  General
                </Button>
                <Button
                  variant={activeSection === 'notifications' ? 'primary' : 'outline-secondary'}
                  className="w-100 mb-2 text-start"
                  onClick={() => setActiveSection('notifications')}
                >
                  <i className="fas fa-bell me-2"></i>
                  Notifications
                </Button>
                <Button
                  variant={activeSection === 'privacy' ? 'primary' : 'outline-secondary'}
                  className="w-100 mb-2 text-start"
                  onClick={() => setActiveSection('privacy')}
                >
                  <i className="fas fa-shield-alt me-2"></i>
                  Privacy & Security
                </Button>
                <Button
                  variant={activeSection === 'appearance' ? 'primary' : 'outline-secondary'}
                  className="w-100 mb-2 text-start"
                  onClick={() => setActiveSection('appearance')}
                >
                  <i className="fas fa-paint-brush me-2"></i>
                  Appearance
                </Button>
                <Button
                  variant="outline-secondary"
                  className="w-100 mb-2 text-start"
                >
                  <i className="fas fa-sync-alt me-2"></i>
                  Sync & Backup
                </Button>
                <Button
                  variant="outline-secondary"
                  className="w-100 text-start"
                >
                  <i className="fas fa-plug me-2"></i>
                  Integrations
                </Button>
              </div>
            </Card.Body>
          </Card>
        </Col>

        <Col md={9}>
          <Card>
            <Card.Body>
              {activeSection === 'general' && (
                <div>
                  <h4 className="mb-4">
                    <i className="fas fa-user me-2"></i>
                    General Settings
                  </h4>
                  <Form>
                    <Row>
                      <Col md={6}>
                        <Form.Group className="mb-3">
                          <Form.Label>Name</Form.Label>
                          <Form.Control
                            type="text"
                            value={formData.general.name}
                            onChange={(e) => handleChange('general', 'name', e.target.value)}
                            placeholder="Enter your name"
                          />
                        </Form.Group>
                      </Col>
                      <Col md={6}>
                        <Form.Group className="mb-3">
                          <Form.Label>Email</Form.Label>
                          <Form.Control
                            type="email"
                            value={formData.general.email}
                            onChange={(e) => handleChange('general', 'email', e.target.value)}
                            placeholder="Enter your email"
                          />
                        </Form.Group>
                      </Col>
                    </Row>
                    <Row>
                      <Col md={6}>
                        <Form.Group className="mb-3">
                          <Form.Label>Language</Form.Label>
                          <Form.Select
                            value={formData.general.language}
                            onChange={(e) => handleChange('general', 'language', e.target.value)}
                          >
                            <option value="en">{getLanguageText('en')}</option>
                            <option value="zh">{getLanguageText('zh')}</option>
                            <option value="es">{getLanguageText('es')}</option>
                            <option value="fr">{getLanguageText('fr')}</option>
                          </Form.Select>
                        </Form.Group>
                      </Col>
                      <Col md={6}>
                        <Form.Group className="mb-3">
                          <Form.Label>Timezone</Form.Label>
                          <Form.Select
                            value={formData.general.timezone}
                            onChange={(e) => handleChange('general', 'timezone', e.target.value)}
                          >
                            <option value="UTC">UTC</option>
                            <option value="America/New_York">Eastern Time</option>
                            <option value="America/Chicago">Central Time</option>
                            <option value="America/Denver">Mountain Time</option>
                            <option value="America/Los_Angeles">Pacific Time</option>
                            <option value="Asia/Shanghai">China Standard Time</option>
                          </Form.Select>
                        </Form.Group>
                      </Col>
                    </Row>
                  </Form>
                </div>
              )}

              {activeSection === 'notifications' && (
                <div>
                  <h4 className="mb-4">
                    <i className="fas fa-bell me-2"></i>
                    Notification Settings
                  </h4>
                  <Form>
                    <div className="mb-4">
                      <h5>Notification Channels</h5>
                      <div className="border rounded p-3">
                        <Form.Check
                          type="switch"
                          id="email-notifications"
                          label={
                            <span>
                              <i className="fas fa-envelope me-2"></i>
                              Email Notifications
                            </span>
                          }
                          checked={formData.notifications.email}
                          onChange={(e) => handleChange('notifications', 'email', e.target.checked)}
                        />
                        <Form.Check
                          type="switch"
                          id="sms-notifications"
                          label={
                            <span>
                              <i className="fas fa-sms me-2"></i>
                              SMS Notifications
                            </span>
                          }
                          checked={formData.notifications.sms}
                          onChange={(e) => handleChange('notifications', 'sms', e.target.checked)}
                          className="mt-2"
                        />
                        <Form.Check
                          type="switch"
                          id="push-notifications"
                          label={
                            <span>
                              <i className="fas fa-bell me-2"></i>
                              Push Notifications
                            </span>
                          }
                          checked={formData.notifications.push}
                          onChange={(e) => handleChange('notifications', 'push', e.target.checked)}
                          className="mt-2"
                        />
                      </div>
                    </div>
                    <div className="mb-4">
                      <h5>Notification Types</h5>
                      <div className="border rounded p-3">
                        <Form.Check
                          type="switch"
                          id="case-updates"
                          label={
                            <span>
                              <i className="fas fa-gavel me-2"></i>
                              Case Updates
                            </span>
                          }
                          checked={formData.notifications.caseUpdates}
                          onChange={(e) => handleChange('notifications', 'caseUpdates', e.target.checked)}
                        />
                        <Form.Check
                          type="switch"
                          id="client-updates"
                          label={
                            <span>
                              <i className="fas fa-users me-2"></i>
                              Client Updates
                            </span>
                          }
                          checked={formData.notifications.clientUpdates}
                          onChange={(e) => handleChange('notifications', 'clientUpdates', e.target.checked)}
                          className="mt-2"
                        />
                        <Form.Check
                          type="switch"
                          id="deadline-reminders"
                          label={
                            <span>
                              <i className="fas fa-clock me-2"></i>
                              Deadline Reminders
                            </span>
                          }
                          checked={formData.notifications.deadlineReminders}
                          onChange={(e) => handleChange('notifications', 'deadlineReminders', e.target.checked)}
                          className="mt-2"
                        />
                        <Form.Check
                          type="switch"
                          id="system-alerts"
                          label={
                            <span>
                              <i className="fas fa-exclamation-triangle me-2"></i>
                              System Alerts
                            </span>
                          }
                          checked={formData.notifications.systemAlerts}
                          onChange={(e) => handleChange('notifications', 'systemAlerts', e.target.checked)}
                          className="mt-2"
                        />
                      </div>
                    </div>
                  </Form>
                </div>
              )}

              {activeSection === 'privacy' && (
                <div>
                  <h4 className="mb-4">
                    <i className="fas fa-shield-alt me-2"></i>
                    Privacy & Security Settings
                  </h4>
                  <Form>
                    <div className="mb-4">
                      <h5>Privacy Controls</h5>
                      <div className="border rounded p-3">
                        <Form.Group className="mb-3">
                          <Form.Label>Profile Visibility</Form.Label>
                          <Form.Select
                            value={formData.privacy.profileVisibility}
                            onChange={(e) => handleChange('privacy', 'profileVisibility', e.target.value)}
                          >
                            <option value="public">{getVisibilityText('public')}</option>
                            <option value="friends">{getVisibilityText('friends')}</option>
                            <option value="private">{getVisibilityText('private')}</option>
                          </Form.Select>
                        </Form.Group>
                        <Form.Group className="mb-3">
                          <Form.Label>Activity Visibility</Form.Label>
                          <Form.Select
                            value={formData.privacy.activityVisibility}
                            onChange={(e) => handleChange('privacy', 'activityVisibility', e.target.value)}
                          >
                            <option value="public">{getVisibilityText('public')}</option>
                            <option value="friends">{getVisibilityText('friends')}</option>
                            <option value="private">{getVisibilityText('private')}</option>
                          </Form.Select>
                        </Form.Group>
                      </div>
                    </div>
                    <div className="mb-4">
                      <h5>Security Settings</h5>
                      <div className="border rounded p-3">
                        <Form.Check
                          type="switch"
                          id="two-factor-auth"
                          label={
                            <span>
                              <i className="fas fa-lock me-2"></i>
                              Two-Factor Authentication
                            </span>
                          }
                          checked={formData.privacy.twoFactorAuth}
                          onChange={(e) => handleChange('privacy', 'twoFactorAuth', e.target.checked)}
                        />
                        <Form.Group className="mt-3">
                          <Form.Label>Session Timeout (minutes)</Form.Label>
                          <Form.Control
                            type="number"
                            min="1"
                            max="120"
                            value={formData.privacy.sessionTimeout}
                            onChange={(e) => handleChange('privacy', 'sessionTimeout', parseInt(e.target.value))}
                          />
                        </Form.Group>
                      </div>
                    </div>
                    <div className="mb-4">
                      <h5>Data Management</h5>
                      <div className="border rounded p-3">
                        <div className="d-grid gap-2">
                          <Button variant="outline-primary">
                            <i className="fas fa-download me-2"></i>
                            Export My Data
                          </Button>
                          <Button variant="outline-warning">
                            <i className="fas fa-history me-2"></i>
                            View Data History
                          </Button>
                          <Button variant="outline-danger">
                            <i className="fas fa-trash me-2"></i>
                            Delete Account
                          </Button>
                        </div>
                      </div>
                    </div>
                  </Form>
                </div>
              )}

              {activeSection === 'appearance' && (
                <div>
                  <h4 className="mb-4">
                    <i className="fas fa-paint-brush me-2"></i>
                    Appearance Settings
                  </h4>
                  <Form>
                    <div className="mb-4">
                      <h5>Theme & Display</h5>
                      <div className="border rounded p-3">
                        <Form.Group className="mb-3">
                          <Form.Label>Theme</Form.Label>
                          <Form.Select
                            value={formData.appearance.theme}
                            onChange={(e) => handleChange('appearance', 'theme', e.target.value)}
                          >
                            <option value="light">{getThemeText('light')}</option>
                            <option value="dark">{getThemeText('dark')}</option>
                            <option value="auto">{getThemeText('auto')}</option>
                          </Form.Select>
                        </Form.Group>
                        <Form.Group className="mb-3">
                          <Form.Label>Font Size</Form.Label>
                          <Form.Select
                            value={formData.appearance.fontSize}
                            onChange={(e) => handleChange('appearance', 'fontSize', e.target.value)}
                          >
                            <option value="small">{getFontSizeText('small')}</option>
                            <option value="medium">{getFontSizeText('medium')}</option>
                            <option value="large">{getFontSizeText('large')}</option>
                          </Form.Select>
                        </Form.Group>
                      </div>
                    </div>
                    <div className="mb-4">
                      <h5>Date & Time Format</h5>
                      <div className="border rounded p-3">
                        <Form.Group className="mb-3">
                          <Form.Label>Date Format</Form.Label>
                          <Form.Select
                            value={formData.appearance.dateFormat}
                            onChange={(e) => handleChange('appearance', 'dateFormat', e.target.value)}
                          >
                            <option value="mm/dd/yyyy">{getDateFormatText('mm/dd/yyyy')}</option>
                            <option value="dd/mm/yyyy">{getDateFormatText('dd/mm/yyyy')}</option>
                            <option value="yyyy-mm-dd">{getDateFormatText('yyyy-mm-dd')}</option>
                          </Form.Select>
                        </Form.Group>
                        <Form.Group className="mb-3">
                          <Form.Label>Time Format</Form.Label>
                          <Form.Select
                            value={formData.appearance.timeFormat}
                            onChange={(e) => handleChange('appearance', 'timeFormat', e.target.value)}
                          >
                            <option value="12h">{getTimeFormatText('12h')}</option>
                            <option value="24h">{getTimeFormatText('24h')}</option>
                          </Form.Select>
                        </Form.Group>
                      </div>
                    </div>
                  </Form>
                </div>
              )}
            </Card.Body>
            <Card.Footer className="d-flex justify-content-end">
              <Button variant="secondary" className="me-2" onClick={handleReset}>
                <i className="fas fa-undo me-2"></i>
                Reset
              </Button>
              <Button variant="primary" onClick={handleSubmit} disabled={saving}>
                {saving ? (
                  <span>
                    <i className="fas fa-spinner fa-spin me-2"></i>
                    Saving...
                  </span>
                ) : (
                  <span>
                    <i className="fas fa-save me-2"></i>
                    Save Settings
                  </span>
                )}
              </Button>
            </Card.Footer>
          </Card>
        </Col>
      </Row>
    </div>
  );
};

export default SettingsPage;