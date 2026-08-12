export type Role = "ADMIN" | "OWNER" | "USER" | "DEMO";

export interface User {
  id: number;
  fullName: string;
  email: string;
  role: Role;
  accountType: string;
  isDemo: boolean;
  createdAt: string;
  updatedAt: string;
}

export type ProjectType = "building" | "infrastructure";

export interface Project {
  id: number;
  userId: number;
  name: string;
  location: string | null;
  type: ProjectType;
  isPitching: boolean;
  isDone: boolean;
  contractValue: number | null;
  timelineMonths: number;
  timelineDays: number;
  createdAt: string;
  updatedAt: string;
  user?: User;
  projectTeam?: ProjectTeamMember[];
  work_items?: ProjectWorkItem[];
  _count?: {
    work_items: number;
    recaps: number;
    invoices: number;
  };
}

export interface ProjectTeamMember {
  id: number;
  projectId: number;
  userId: number;
  role: "owner" | "editor" | "viewer";
  createdAt: string;
  user?: User;
  project?: Project;
}

export interface ProjectWorkItem {
  id: number;
  description: string;
  volume: number;
  unit: string;
  unitPrice: number;
  totalCost: number;
  category: WorkCategory;
}

export type WorkCategory =
  | "preparation"
  | "foundation"
  | "concrete"
  | "canopy"
  | "steel"
  | "stairs"
  | "roof"
  | "wall"
  | "plastering"
  | "finishing"
  | "tiles"
  | "paving"
  | "painting"
  | "doors"
  | "interior"
  | "toilet"
  | "mep"
  | "custom";

export type CalculationMethod =
  | "ahsp"
  | "manual"
  | "lump_sum"
  | "manual_price"
  | "custom_price";

export type ComponentType = "material" | "labor" | "equipment";

export interface WorkItem {
  id: number;
  projectId: number;
  category: WorkCategory;
  description: string;
  volume: number;
  unit: string;
  unitPrice: number;
  totalCost: number;
  calculationMethod: CalculationMethod;
  level: string | null;
  type: string | null;
  analysisMasterId: number | null;
  duration: number | null;
  totalDuration: number | null;
  createdAt: string;
  updatedAt: string;
  itemDetails?: WorkItemDetail[];
}

export interface WorkItemDetail {
  id: number;
  workItemId: number;
  priceMasterId: number | null;
  name: string;
  unit: string;
  coefficient: number;
  unitPrice: number;
  totalCost: number;
  type: ComponentType;
}

export interface AnalysisMaster {
  id: number;
  code: string;
  name: string;
  level: number;
  parentId: number | null;
  unit: string | null;
  unitPrice?: number | null;
  isGlobal: boolean;
  isSystem?: boolean;
  ahspCode?: string | null;
  generalCost?: number;
  userId: number | null;
  createdAt: string;
  children?: AnalysisMaster[];
}

export interface AnalysisComponent {
  id: number;
  analysisMasterId: number;
  componentId: number | null;
  coefficient: number;
  type: ComponentType;
  name: string | null;
  unit: string | null;
  unitPrice: number | null;
  totalPrice: number | null;
  referenceCode: string | null;
  duration: number | null;
  sequence: number;
}

export interface PriceMaster {
  id: number;
  name: string;
  unit: string;
  price: number;
  type: ComponentType;
  isGlobal: boolean;
  userId: number | null;
  createdAt: string;
  updatedAt: string;
}

export interface Recap {
  id: number;
  projectId: number;
  category: string;
  description: string;
  sequence: number;
  margin: number | null;
  createdAt: string;
  updatedAt: string;
}

export interface InvoiceItem {
  id?: number;
  description: string;
  qty: number;
  unit: string;
  unitPrice: number;
}

export interface Invoice {
  id: number;
  projectId: number;
  number: string;
  date: string;
  dueDate?: string | null;
  poNumber?: string | null;
  buyerName?: string | null;
  buyerAddress?: string | null;
  buyerContact?: string | null;
  discount: number;
  taxRate: number;
  paymentBank?: string | null;
  paymentAccountNumber?: string | null;
  paymentAccountName?: string | null;
  notes?: string | null;
  financeName?: string | null;
  total: number;
  status: "draft" | "sent" | "paid";
  items?: InvoiceItem[];
  createdAt: string;
}

export interface Logistics {
  id: number;
  projectId: number;
  materialName: string;
  unit: string;
  volume: number;
  unitPrice: number;
  totalCost: number;
  date: string | null;
  description: string | null;
  createdAt: string;
}

export interface Transaction {
  id: number;
  projectId: number;
  date: string;
  category: string;
  amount: number;
  description: string | null;
  type: "expense" | "income";
  status: "draft" | "approved" | "reverted";
  logisticsId: number | null;
  invoiceId: number | null;
  createdAt: string;
}

export interface News {
  id: number;
  title: string;
  content: string;
  isActive: boolean;
  createdAt: string;
}

export interface CalculationResult {
  volume: number;
  unitPrice: number;
  totalCost: number;
  breakdown: {
    material: number;
    labor: number;
    equipment: number;
  };
}

export interface RABResult {
  subtotal: number;
  overhead: number;
  profit: number;
  ppn: number;
  total: number;
}
