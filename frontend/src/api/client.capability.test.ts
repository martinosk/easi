import { describe, it, expect, vi, beforeEach } from 'vitest';
import axios from 'axios';
import type { MockedFunction } from 'vitest';
import type {
  Capability,
  CapabilityDependency,
  CapabilityRealization,
  CollectionResponse,
} from './types';

vi.mock('axios');

const mockAxiosInstance = {
  get: vi.fn(),
  post: vi.fn(),
  put: vi.fn(),
  patch: vi.fn(),
  delete: vi.fn(),
  interceptors: {
    response: {
      use: vi.fn(),
    },
  },
};

(axios.create as MockedFunction<typeof axios.create>).mockReturnValue(mockAxiosInstance as any);

describe('API Client - Capability Operations', () => {
  let apiClient: typeof import('./client').apiClient;
  let responseInterceptorError: (error: any) => never;

  beforeEach(async () => {
    vi.clearAllMocks();
    vi.resetModules();

    mockAxiosInstance.interceptors.response.use.mockImplementation(
      (_onFulfilled: any, onRejected: any) => {
        responseInterceptorError = onRejected;
      }
    );

    const clientModule = await import('./client');
    apiClient = clientModule.apiClient;
  });

  describe('getCapabilities', () => {
    it('should return all capabilities from collection response', async () => {
      const mockCapabilities: Capability[] = [
        {
          id: 'cap-1',
          name: 'Customer Management',
          level: 'L1',
          createdAt: '2024-01-01T00:00:00Z',
          _links: { self: { href: '/api/v1/capabilities/cap-1' } },
        },
        {
          id: 'cap-2',
          name: 'Order Processing',
          level: 'L2',
          parentId: 'cap-1',
          createdAt: '2024-01-01T00:00:00Z',
          _links: { self: { href: '/api/v1/capabilities/cap-2' } },
        },
      ];

      const response: CollectionResponse<Capability> = {
        data: mockCapabilities,
        _links: { self: { href: '/api/v1/capabilities' } },
      };

      mockAxiosInstance.get.mockResolvedValueOnce({ data: response });

      const result = await apiClient.getCapabilities();

      expect(mockAxiosInstance.get).toHaveBeenCalledWith('/api/v1/capabilities');
      expect(result).toEqual(mockCapabilities);
    });

    it('should return empty array when data is null', async () => {
      const response = {
        data: null,
        _links: { self: { href: '/api/v1/capabilities' } },
      };

      mockAxiosInstance.get.mockResolvedValueOnce({ data: response });

      const result = await apiClient.getCapabilities();

      expect(result).toEqual([]);
    });
  });

  describe('getCapabilityById', () => {
    it('should return capability by id', async () => {
      const mockCapability: Capability = {
        id: 'cap-1',
        name: 'Customer Management',
        description: 'Manage customer data',
        level: 'L1',
        strategyPillar: 'Growth',
        maturityLevel: 'Optimized',
        status: 'Active',
        experts: [
          {
            name: 'John Doe',
            role: 'Solution Architect',
            contact: 'john.doe@example.com',
            addedAt: '2024-01-01T00:00:00Z',
          },
        ],
        tags: ['core', 'customer'],
        createdAt: '2024-01-01T00:00:00Z',
        _links: { self: { href: '/api/v1/capabilities/cap-1' } },
      };

      mockAxiosInstance.get.mockResolvedValueOnce({ data: mockCapability });

      const result = await apiClient.getCapabilityById('cap-1');

      expect(mockAxiosInstance.get).toHaveBeenCalledWith('/api/v1/capabilities/cap-1');
      expect(result).toEqual(mockCapability);
    });
  });

  describe('getCapabilityChildren', () => {
    it('should return children of a capability', async () => {
      const mockChildren: Capability[] = [
        {
          id: 'cap-2',
          name: 'Customer Onboarding',
          level: 'L2',
          parentId: 'cap-1',
          createdAt: '2024-01-01T00:00:00Z',
          _links: { self: { href: '/api/v1/capabilities/cap-2' } },
        },
      ];

      const response: CollectionResponse<Capability> = {
        data: mockChildren,
        _links: { self: { href: '/api/v1/capabilities/cap-1/children' } },
      };

      mockAxiosInstance.get.mockResolvedValueOnce({ data: response });

      const result = await apiClient.getCapabilityChildren('cap-1');

      expect(mockAxiosInstance.get).toHaveBeenCalledWith('/api/v1/capabilities/cap-1/children');
      expect(result).toEqual(mockChildren);
    });
  });

  describe('createCapability', () => {
    it('should create a new L1 capability', async () => {
      const request = {
        name: 'New Capability',
        description: 'A new capability',
        level: 'L1' as const,
      };

      const mockCapability: Capability = {
        id: 'cap-new',
        name: 'New Capability',
        description: 'A new capability',
        level: 'L1',
        createdAt: '2024-01-01T00:00:00Z',
        _links: { self: { href: '/api/v1/capabilities/cap-new' } },
      };

      mockAxiosInstance.post.mockResolvedValueOnce({ data: mockCapability });

      const result = await apiClient.createCapability(request);

      expect(mockAxiosInstance.post).toHaveBeenCalledWith('/api/v1/capabilities', request);
      expect(result).toEqual(mockCapability);
    });

    it('should create a child capability with parentId', async () => {
      const request = {
        name: 'Child Capability',
        description: 'A child capability',
        parentId: 'cap-1',
        level: 'L2' as const,
      };

      const mockCapability: Capability = {
        id: 'cap-child',
        name: 'Child Capability',
        description: 'A child capability',
        level: 'L2',
        parentId: 'cap-1',
        createdAt: '2024-01-01T00:00:00Z',
        _links: { self: { href: '/api/v1/capabilities/cap-child' } },
      };

      mockAxiosInstance.post.mockResolvedValueOnce({ data: mockCapability });

      const result = await apiClient.createCapability(request);

      expect(mockAxiosInstance.post).toHaveBeenCalledWith('/api/v1/capabilities', request);
      expect(result).toEqual(mockCapability);
    });
  });

  describe('updateCapability', () => {
    it('should update capability name and description', async () => {
      const request = {
        name: 'Updated Capability',
        description: 'Updated description',
      };

      const mockCapability: Capability = {
        id: 'cap-1',
        name: 'Updated Capability',
        description: 'Updated description',
        level: 'L1',
        createdAt: '2024-01-01T00:00:00Z',
        _links: { self: { href: '/api/v1/capabilities/cap-1' } },
      };

      mockAxiosInstance.put.mockResolvedValueOnce({ data: mockCapability });

      const result = await apiClient.updateCapability('cap-1', request);

      expect(mockAxiosInstance.put).toHaveBeenCalledWith('/api/v1/capabilities/cap-1', request);
      expect(result).toEqual(mockCapability);
    });
  });

  describe('updateCapabilityMetadata', () => {
    it('should update capability metadata', async () => {
      const request = {
        strategyPillar: 'Growth',
        pillarWeight: 0.75,
        maturityLevel: 'Optimized',
        ownershipModel: 'Centralized',
        primaryOwner: 'John Doe',
        eaOwner: 'Jane Smith',
        status: 'Active',
      };

      const mockCapability: Capability = {
        id: 'cap-1',
        name: 'Customer Management',
        level: 'L1',
        ...request,
        createdAt: '2024-01-01T00:00:00Z',
        _links: { self: { href: '/api/v1/capabilities/cap-1' } },
      };

      mockAxiosInstance.put.mockResolvedValueOnce({ data: mockCapability });

      const result = await apiClient.updateCapabilityMetadata('cap-1', request);

      expect(mockAxiosInstance.put).toHaveBeenCalledWith('/api/v1/capabilities/cap-1/metadata', request);
      expect(result).toEqual(mockCapability);
    });
  });

  describe('addCapabilityExpert', () => {
    it('should add an expert to a capability', async () => {
      const request = {
        expertName: 'John Doe',
        expertRole: 'Solution Architect',
        contactInfo: 'john.doe@example.com',
      };

      mockAxiosInstance.post.mockResolvedValueOnce({ data: undefined });

      await apiClient.addCapabilityExpert('cap-1', request);

      expect(mockAxiosInstance.post).toHaveBeenCalledWith('/api/v1/capabilities/cap-1/experts', request);
    });
  });

  describe('addCapabilityTag', () => {
    it('should add a tag to a capability', async () => {
      const request = { tag: 'core' };

      mockAxiosInstance.post.mockResolvedValueOnce({ data: undefined });

      await apiClient.addCapabilityTag('cap-1', request);

      expect(mockAxiosInstance.post).toHaveBeenCalledWith('/api/v1/capabilities/cap-1/tags', request);
    });
  });
});

describe('API Client - Capability Dependency Operations', () => {
  let apiClient: typeof import('./client').apiClient;

  beforeEach(async () => {
    vi.clearAllMocks();
    vi.resetModules();

    mockAxiosInstance.interceptors.response.use.mockImplementation(() => {});

    const clientModule = await import('./client');
    apiClient = clientModule.apiClient;
  });

  describe('getCapabilityDependencies', () => {
    it('should return all capability dependencies', async () => {
      const mockDependencies: CapabilityDependency[] = [
        {
          id: 'dep-1',
          sourceCapabilityId: 'cap-1',
          targetCapabilityId: 'cap-2',
          dependencyType: 'Requires',
          createdAt: '2024-01-01T00:00:00Z',
          _links: { self: { href: '/api/v1/capability-dependencies/dep-1' } },
        },
      ];

      const response: CollectionResponse<CapabilityDependency> = {
        data: mockDependencies,
        _links: { self: { href: '/api/v1/capability-dependencies' } },
      };

      mockAxiosInstance.get.mockResolvedValueOnce({ data: response });

      const result = await apiClient.getCapabilityDependencies();

      expect(mockAxiosInstance.get).toHaveBeenCalledWith('/api/v1/capability-dependencies');
      expect(result).toEqual(mockDependencies);
    });
  });

  describe('getOutgoingDependencies', () => {
    it('should return outgoing dependencies for a capability', async () => {
      const mockDependencies: CapabilityDependency[] = [
        {
          id: 'dep-1',
          sourceCapabilityId: 'cap-1',
          targetCapabilityId: 'cap-2',
          dependencyType: 'Requires',
          createdAt: '2024-01-01T00:00:00Z',
          _links: { self: { href: '/api/v1/capability-dependencies/dep-1' } },
        },
      ];

      const response: CollectionResponse<CapabilityDependency> = {
        data: mockDependencies,
        _links: { self: { href: '/api/v1/capabilities/cap-1/dependencies/outgoing' } },
      };

      mockAxiosInstance.get.mockResolvedValueOnce({ data: response });

      const result = await apiClient.getOutgoingDependencies('cap-1');

      expect(mockAxiosInstance.get).toHaveBeenCalledWith('/api/v1/capabilities/cap-1/dependencies/outgoing');
      expect(result).toEqual(mockDependencies);
    });
  });

  describe('getIncomingDependencies', () => {
    it('should return incoming dependencies for a capability', async () => {
      const mockDependencies: CapabilityDependency[] = [
        {
          id: 'dep-2',
          sourceCapabilityId: 'cap-2',
          targetCapabilityId: 'cap-1',
          dependencyType: 'Enables',
          createdAt: '2024-01-01T00:00:00Z',
          _links: { self: { href: '/api/v1/capability-dependencies/dep-2' } },
        },
      ];

      const response: CollectionResponse<CapabilityDependency> = {
        data: mockDependencies,
        _links: { self: { href: '/api/v1/capabilities/cap-1/dependencies/incoming' } },
      };

      mockAxiosInstance.get.mockResolvedValueOnce({ data: response });

      const result = await apiClient.getIncomingDependencies('cap-1');

      expect(mockAxiosInstance.get).toHaveBeenCalledWith('/api/v1/capabilities/cap-1/dependencies/incoming');
      expect(result).toEqual(mockDependencies);
    });
  });

  describe('createCapabilityDependency', () => {
    it('should create a new capability dependency', async () => {
      const request = {
        sourceCapabilityId: 'cap-1',
        targetCapabilityId: 'cap-2',
        dependencyType: 'Requires' as const,
        description: 'Needs customer data',
      };

      const mockDependency: CapabilityDependency = {
        id: 'dep-new',
        sourceCapabilityId: 'cap-1',
        targetCapabilityId: 'cap-2',
        dependencyType: 'Requires',
        description: 'Needs customer data',
        createdAt: '2024-01-01T00:00:00Z',
        _links: { self: { href: '/api/v1/capability-dependencies/dep-new' } },
      };

      mockAxiosInstance.post.mockResolvedValueOnce({ data: mockDependency });

      const result = await apiClient.createCapabilityDependency(request);

      expect(mockAxiosInstance.post).toHaveBeenCalledWith('/api/v1/capability-dependencies', request);
      expect(result).toEqual(mockDependency);
    });
  });

  describe('deleteCapabilityDependency', () => {
    it('should delete a capability dependency', async () => {
      mockAxiosInstance.delete.mockResolvedValueOnce({ data: undefined });

      await apiClient.deleteCapabilityDependency('dep-1');

      expect(mockAxiosInstance.delete).toHaveBeenCalledWith('/api/v1/capability-dependencies/dep-1');
    });
  });
});

describe('API Client - Capability Realization Operations', () => {
  let apiClient: typeof import('./client').apiClient;

  beforeEach(async () => {
    vi.clearAllMocks();
    vi.resetModules();

    mockAxiosInstance.interceptors.response.use.mockImplementation(() => {});

    const clientModule = await import('./client');
    apiClient = clientModule.apiClient;
  });

  describe('getSystemsByCapability', () => {
    it('should return systems linked to a capability', async () => {
      const mockRealizations: CapabilityRealization[] = [
        {
          id: 'real-1',
          capabilityId: 'cap-1',
          componentId: 'comp-1',
          realizationLevel: 'Full',
          notes: 'Fully implements customer management',
          linkedAt: '2024-01-01T00:00:00Z',
          _links: { self: { href: '/api/v1/capability-realizations/real-1' } },
        },
      ];

      const response: CollectionResponse<CapabilityRealization> = {
        data: mockRealizations,
        _links: { self: { href: '/api/v1/capabilities/cap-1/systems' } },
      };

      mockAxiosInstance.get.mockResolvedValueOnce({ data: response });

      const result = await apiClient.getSystemsByCapability('cap-1');

      expect(mockAxiosInstance.get).toHaveBeenCalledWith('/api/v1/capabilities/cap-1/systems');
      expect(result).toEqual(mockRealizations);
    });
  });

  describe('getCapabilitiesByComponent', () => {
    it('should return capabilities realized by a component', async () => {
      const mockRealizations: CapabilityRealization[] = [
        {
          id: 'real-1',
          capabilityId: 'cap-1',
          componentId: 'comp-1',
          realizationLevel: 'Full',
          linkedAt: '2024-01-01T00:00:00Z',
          _links: { self: { href: '/api/v1/capability-realizations/real-1' } },
        },
        {
          id: 'real-2',
          capabilityId: 'cap-2',
          componentId: 'comp-1',
          realizationLevel: 'Partial',
          linkedAt: '2024-01-01T00:00:00Z',
          _links: { self: { href: '/api/v1/capability-realizations/real-2' } },
        },
      ];

      const response: CollectionResponse<CapabilityRealization> = {
        data: mockRealizations,
        _links: { self: { href: '/api/v1/capability-realizations/by-component/comp-1' } },
      };

      mockAxiosInstance.get.mockResolvedValueOnce({ data: response });

      const result = await apiClient.getCapabilitiesByComponent('comp-1');

      expect(mockAxiosInstance.get).toHaveBeenCalledWith('/api/v1/capability-realizations/by-component/comp-1');
      expect(result).toEqual(mockRealizations);
    });
  });

  describe('linkSystemToCapability', () => {
    it('should link a system to a capability', async () => {
      const request = {
        componentId: 'comp-1',
        realizationLevel: 'Full' as const,
        notes: 'Primary system for this capability',
      };

      const mockRealization: CapabilityRealization = {
        id: 'real-new',
        capabilityId: 'cap-1',
        componentId: 'comp-1',
        realizationLevel: 'Full',
        notes: 'Primary system for this capability',
        linkedAt: '2024-01-01T00:00:00Z',
        _links: { self: { href: '/api/v1/capability-realizations/real-new' } },
      };

      mockAxiosInstance.post.mockResolvedValueOnce({ data: mockRealization });

      const result = await apiClient.linkSystemToCapability('cap-1', request);

      expect(mockAxiosInstance.post).toHaveBeenCalledWith('/api/v1/capabilities/cap-1/systems', request);
      expect(result).toEqual(mockRealization);
    });
  });

  describe('updateRealization', () => {
    it('should update a realization', async () => {
      const request = {
        realizationLevel: 'Partial' as const,
        notes: 'Updated notes',
      };

      const mockRealization: CapabilityRealization = {
        id: 'real-1',
        capabilityId: 'cap-1',
        componentId: 'comp-1',
        realizationLevel: 'Partial',
        notes: 'Updated notes',
        linkedAt: '2024-01-01T00:00:00Z',
        _links: { self: { href: '/api/v1/capability-realizations/real-1' } },
      };

      mockAxiosInstance.put.mockResolvedValueOnce({ data: mockRealization });

      const result = await apiClient.updateRealization('real-1', request);

      expect(mockAxiosInstance.put).toHaveBeenCalledWith('/api/v1/capability-realizations/real-1', request);
      expect(result).toEqual(mockRealization);
    });
  });

  describe('deleteRealization', () => {
    it('should delete a realization', async () => {
      mockAxiosInstance.delete.mockResolvedValueOnce({ data: undefined });

      await apiClient.deleteRealization('real-1');

      expect(mockAxiosInstance.delete).toHaveBeenCalledWith('/api/v1/capability-realizations/real-1');
    });
  });
});

describe('API Client - Error Handling', () => {
  let ApiError: typeof import('./types').ApiError;
  let responseInterceptorError: (error: any) => never;

  beforeEach(async () => {
    vi.clearAllMocks();
    vi.resetModules();

    mockAxiosInstance.interceptors.response.use.mockImplementation(
      (_onFulfilled: any, onRejected: any) => {
        responseInterceptorError = onRejected;
      }
    );

    const typesModule = await import('./types');
    ApiError = typesModule.ApiError;
    await import('./client');
  });

  it('should throw ApiError with message from response', () => {
    const axiosError = {
      response: {
        status: 400,
        data: {
          message: 'Capability name is required',
        },
      },
    };

    expect(() => responseInterceptorError(axiosError)).toThrow();
    try {
      responseInterceptorError(axiosError);
    } catch (error) {
      expect(error).toBeInstanceOf(ApiError);
      expect((error as typeof ApiError.prototype).message).toBe('Capability name is required');
      expect((error as typeof ApiError.prototype).statusCode).toBe(400);
    }
  });

  it('should throw ApiError with error field when message is absent', () => {
    const axiosError = {
      response: {
        status: 409,
        data: {
          error: 'Conflict',
        },
      },
    };

    try {
      responseInterceptorError(axiosError);
    } catch (error) {
      expect(error).toBeInstanceOf(ApiError);
      expect((error as typeof ApiError.prototype).message).toBe('Conflict');
      expect((error as typeof ApiError.prototype).statusCode).toBe(409);
    }
  });

  it('should throw ApiError with details when present', () => {
    const axiosError = {
      response: {
        status: 400,
        data: {
          details: {
            name: 'Name must be at least 3 characters',
            level: 'Invalid level',
          },
        },
      },
    };

    try {
      responseInterceptorError(axiosError);
    } catch (error) {
      expect(error).toBeInstanceOf(ApiError);
      expect((error as typeof ApiError.prototype).message).toContain('Name must be at least 3 characters');
      expect((error as typeof ApiError.prototype).statusCode).toBe(400);
    }
  });

  it('should throw ApiError with fallback message when no error info', () => {
    const axiosError = {
      response: {
        status: 500,
        data: {},
      },
    };

    try {
      responseInterceptorError(axiosError);
    } catch (error) {
      expect(error).toBeInstanceOf(ApiError);
      expect((error as typeof ApiError.prototype).message).toBe('An error occurred');
      expect((error as typeof ApiError.prototype).statusCode).toBe(500);
    }
  });

  it('should use axios error message when response is missing', () => {
    const axiosError = {
      message: 'Network Error',
      response: undefined,
    };

    try {
      responseInterceptorError(axiosError);
    } catch (error) {
      expect(error).toBeInstanceOf(ApiError);
      expect((error as typeof ApiError.prototype).message).toBe('Network Error');
      expect((error as typeof ApiError.prototype).statusCode).toBe(500);
    }
  });

  it('should handle 404 not found for capability', () => {
    const axiosError = {
      response: {
        status: 404,
        data: {
          message: 'Capability not found',
        },
      },
    };

    try {
      responseInterceptorError(axiosError);
    } catch (error) {
      expect(error).toBeInstanceOf(ApiError);
      expect((error as typeof ApiError.prototype).statusCode).toBe(404);
    }
  });

  it('should handle 409 conflict for duplicate capability', () => {
    const axiosError = {
      response: {
        status: 409,
        data: {
          message: 'Capability with this name already exists',
        },
      },
    };

    try {
      responseInterceptorError(axiosError);
    } catch (error) {
      expect(error).toBeInstanceOf(ApiError);
      expect((error as typeof ApiError.prototype).message).toBe('Capability with this name already exists');
      expect((error as typeof ApiError.prototype).statusCode).toBe(409);
    }
  });
});
