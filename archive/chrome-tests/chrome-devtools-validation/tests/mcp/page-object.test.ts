import { PageObject } from '../../src/mcp/page-object';
import { ChromeDevToolsService } from '../../src/mcp/devtools-service';
import { DevToolsElement } from '../../src/types';
import { jest } from '@jest/globals';

// Create a test PageObject implementation
class TestPageObject extends PageObject {
  constructor(service: ChromeDevToolsService) {
    super(service);
    this.url = '/test';
  }

  async testMethod(): Promise<string> {
    return 'test-result';
  }
}

// Mock ChromeDevToolsService
jest.mock('../../src/mcp/devtools-service');
const MockChromeDevToolsService = require('../../src/mcp/devtools-service').ChromeDevToolsService;

// Mock setTimeout to speed up tests
jest.spyOn(global, 'setTimeout').mockImplementation((callback: Function, _delay?: number) => {
  if (callback) callback();
  return {} as NodeJS.Timeout;
});

describe('PageObject', () => {
  let pageObject: TestPageObject;
  let mockService: any;

  beforeEach(() => {
    mockService = {
      navigate: jest.fn(),
      wait: jest.fn(),
      getElement: jest.fn(),
      click: jest.fn(),
      fill: jest.fn(),
      select: jest.fn(),
      executeScript: jest.fn(),
      screenshot: jest.fn(),
    };

    // Mock the ChromeDevToolsService constructor
    MockChromeDevToolsService.mockImplementation(() => mockService);

    pageObject = new TestPageObject(mockService);
  });

  afterEach(() => {
    jest.clearAllMocks();
  });

  describe('constructor', () => {
    it('should create PageObject with service', () => {
      expect(pageObject).toBeInstanceOf(PageObject);
      expect(pageObject).toBeInstanceOf(TestPageObject);
    });
  });

  describe('navigate', () => {
    it('should navigate to page URL', async () => {
      mockService.navigate.mockResolvedValue(undefined);
      mockService.wait.mockResolvedValue(undefined);

      await pageObject.navigate();

      expect(mockService.navigate).toHaveBeenCalledWith('/test');
      expect(mockService.wait).toHaveBeenCalledWith({
        condition: 'navigation',
        timeout: 30000,
      });
    });

    it('should throw error when URL is not defined', async () => {
      // Create a PageObject without URL
      class NoUrlPageObject extends PageObject {
        constructor(service: any) {
          super(service);
          // No URL set
        }
      }

      const noUrlPage = new NoUrlPageObject(mockService);

      await expect(noUrlPage.navigate()).rejects.toThrow('NoUrlPageObject 未定义URL');
    });
  });

  describe('getElement', () => {
    it('should delegate to service getElement', async () => {
      const mockElement: DevToolsElement = {
        uid: 'test-uid',
        tagName: 'div',
        attributes: {},
        visible: true,
        enabled: true,
        x: 0,
        y: 0,
        width: 100,
        height: 50,
      };

      mockService.getElement.mockResolvedValue(mockElement);

      const result = await pageObject['getElement']('#test-element');

      expect(result).toBe(mockElement);
      expect(mockService.getElement).toHaveBeenCalledWith('#test-element');
    });
  });

  describe('waitForElement', () => {
    it('should wait for element to become visible', async () => {
      const visibleElement: DevToolsElement = {
        uid: 'test-uid',
        tagName: 'div',
        attributes: {},
        visible: true,
        enabled: true,
        x: 0,
        y: 0,
        width: 100,
        height: 50,
      };

      // First call returns null, second returns visible element
      mockService.getElement
        .mockResolvedValueOnce(null)
        .mockResolvedValueOnce(visibleElement);

      const result = await pageObject['waitForElement']('#test-element', 1000);

      expect(result).toBe(visibleElement);
      expect(mockService.getElement).toHaveBeenCalledTimes(2);
    });

    it('should timeout if element never becomes visible', async () => {
      mockService.getElement.mockResolvedValue(null);

      await expect(
        pageObject['waitForElement']('#test-element', 100)
      ).rejects.toThrow('元素 #test-element 在 100ms 内未变为可见');
    });
  });

  describe('waitForElementHidden', () => {
    it('should wait for element to become hidden', async () => {
      // First call returns visible element, second returns null
      const visibleElement: DevToolsElement = {
        uid: 'test-uid',
        tagName: 'div',
        attributes: {},
        visible: true,
        enabled: true,
        x: 0,
        y: 0,
        width: 100,
        height: 50,
      };

      mockService.getElement
        .mockResolvedValueOnce(visibleElement)
        .mockResolvedValueOnce(null);

      await expect(
        pageObject['waitForElementHidden']('#test-element', 1000)
      ).resolves.not.toThrow();
    });

    it('should pass immediately if element is already hidden', async () => {
      mockService.getElement.mockResolvedValue(null);

      await expect(
        pageObject['waitForElementHidden']('#test-element', 1000)
      ).resolves.not.toThrow();

      expect(mockService.getElement).toHaveBeenCalledTimes(1);
    });
  });

  describe('click', () => {
    it('should click element without waiting for navigation', async () => {
      mockService.click.mockResolvedValue(undefined);

      await pageObject['click']('#test-button', false);

      expect(mockService.click).toHaveBeenCalledWith('#test-button');
    });

    it('should wait for navigation when specified', async () => {
      mockService.click.mockResolvedValue(undefined);
      mockService.wait.mockResolvedValue(undefined);

      await pageObject['click']('#test-button', true);

      expect(mockService.click).toHaveBeenCalledWith('#test-button');
      expect(mockService.wait).toHaveBeenCalledWith({
        condition: 'navigation',
        timeout: 30000,
      });
    });
  });

  describe('fill', () => {
    it('should fill input field', async () => {
      mockService.fill.mockResolvedValue(undefined);

      await pageObject['fill']('#username', 'testuser');

      expect(mockService.fill).toHaveBeenCalledWith('#username', 'testuser', {
        clear: true,
        delay: 0,
      });
    });

    it('should fill with clear option', async () => {
      mockService.fill.mockResolvedValue(undefined);

      await pageObject['fill']('#username', 'testuser', false);

      expect(mockService.fill).toHaveBeenCalledWith('#username', 'testuser', {
        clear: false,
        delay: 0,
      });
    });
  });

  describe('select', () => {
    it('should select dropdown options', async () => {
      mockService.select.mockResolvedValue(undefined);

      await pageObject['select']('#country', ['China', 'USA']);

      expect(mockService.select).toHaveBeenCalledWith('#country', ['China', 'USA'], {
        force: false,
      });
    });
  });

  describe('text operations', () => {
    it('should get element text', async () => {
      mockService.executeScript.mockResolvedValue('Test Text');

      const result = await pageObject['getText']('#test-element');

      expect(result).toBe('Test Text');
      expect(mockService.executeScript).toHaveBeenCalledWith(
        expect.stringContaining('document.querySelector'),
      );
    });

    it('should get element attribute', async () => {
      mockService.executeScript.mockResolvedValue('test-value');

      const result = await pageObject['getAttribute']('#test-element', 'data-test');

      expect(result).toBe('test-value');
      expect(mockService.executeScript).toHaveBeenCalledWith(
        expect.stringContaining('getAttribute'),
      );
    });
  });

  describe('visibility checks', () => {
    it('should check element visibility', async () => {
      mockService.executeScript.mockResolvedValue(true);

      const result = await pageObject['isVisible']('#test-element');

      expect(result).toBe(true);
      expect(mockService.executeScript).toHaveBeenCalledWith(
        expect.stringContaining('getComputedStyle'),
      );
    });

    it('should check element existence', async () => {
      mockService.executeScript.mockResolvedValue(true);

      const result = await pageObject['exists']('#test-element');

      expect(result).toBe(true);
      expect(mockService.executeScript).toHaveBeenCalledWith(
        expect.stringContaining('querySelector'),
      );
    });
  });

  describe('value operations', () => {
    it('should get element value', async () => {
      mockService.executeScript.mockResolvedValue('test-value');

      const result = await pageObject['getValue']('#input-field');

      expect(result).toBe('test-value');
    });

    it('should set element value', async () => {
      mockService.executeScript.mockResolvedValue(undefined);

      await pageObject['setValue']('#input-field', 'new-value');

      expect(mockService.executeScript).toHaveBeenCalledWith(
        expect.stringContaining('value'),
      );
    });
  });

  describe('page operations', () => {
    it('should get page title', async () => {
      mockService.executeScript.mockResolvedValue('Test Page');

      const result = await pageObject.getTitle();

      expect(result).toBe('Test Page');
      expect(mockService.executeScript).toHaveBeenCalledWith('return document.title;');
    });

    it('should get page URL', async () => {
      mockService.executeScript.mockResolvedValue('https://example.com');

      const result = await pageObject.getUrl();

      expect(result).toBe('https://example.com');
      expect(mockService.executeScript).toHaveBeenCalledWith('return window.location.href;');
    });

    it('should refresh page', async () => {
      mockService.executeScript.mockResolvedValue(undefined);
      mockService.wait.mockResolvedValue(undefined);

      await pageObject.refresh();

      expect(mockService.executeScript).toHaveBeenCalledWith('window.location.reload();');
      expect(mockService.wait).toHaveBeenCalledWith({
        condition: 'navigation',
        timeout: 30000,
      });
    });
  });

  describe('expectations', () => {
    beforeEach(() => {
      mockService.executeScript.mockImplementation((script: string) => {
        if (script.includes('location.href')) {
          return 'https://example.com/dashboard';
        }
        return '';
      });
    });

    it('should validate URL contains pattern', async () => {
      await expect(
        pageObject['expectUrl']('dashboard')
      ).resolves.not.toThrow();
    });

    it('should validate URL matches regex', async () => {
      await expect(
        pageObject['expectUrl'](/dashboard/)
      ).resolves.not.toThrow();
    });

    it('should throw error when URL does not contain pattern', async () => {
      mockService.executeScript.mockResolvedValue('https://example.com/login');

      await expect(
        pageObject['expectUrl']('dashboard')
      ).rejects.toThrow('期望URL包含 \'dashboard\'');
    });

    it('should validate text contains', async () => {
      mockService.executeScript.mockResolvedValue('Welcome to the dashboard');

      await expect(
        pageObject['expectTextContains']('#welcome', 'Welcome')
      ).resolves.not.toThrow();
    });

    it('should throw error when text does not contain expected', async () => {
      mockService.executeScript.mockResolvedValue('Goodbye');

      await expect(
        pageObject['expectTextContains']('#welcome', 'Welcome')
      ).rejects.toThrow('期望元素 \'#welcome\' 包含文本 \'Welcome\'');
    });
  });

  describe('screenshot', () => {
    it('should take screenshot', async () => {
      const mockScreenshot = 'base64-image-data';
      mockService.screenshot.mockResolvedValue(mockScreenshot);

      const result = await pageObject['screenshot']('test-screenshot');

      expect(result).toBe(mockScreenshot);
      expect(mockService.screenshot).toHaveBeenCalledWith({
        filename: 'test-screenshot',
      });
    });
  });

  describe('waitFor', () => {
    it('should wait for condition to be true', async () => {
      let attempts = 0;
      const condition = jest.fn().mockImplementation(async () => {
        attempts++;
        return attempts >= 2; // Returns true on second attempt
      }) as jest.MockedFunction<() => Promise<boolean>>;

      await expect(
        pageObject['waitFor'](condition, 1000, 100)
      ).resolves.not.toThrow();

      expect(condition).toHaveBeenCalledTimes(2);
    });

    it('should timeout if condition never becomes true', async () => {
      const condition = jest.fn().mockImplementation(async () => false) as jest.MockedFunction<() => Promise<boolean>>;

      await expect(
        pageObject['waitFor'](condition, 100, 10)
      ).rejects.toThrow('条件在 100ms 内未满足');
    });
  });
});