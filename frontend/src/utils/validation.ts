import { z } from 'zod';

// Login schema
export const loginSchema = z.object({
  email: z
    .string()
    .min(1, 'auth.emailRequired')
    .email('auth.emailInvalid'),
  password: z
    .string()
    .min(1, 'auth.passwordRequired')
    .min(8, 'auth.passwordMinLength'),
});

export type LoginFormData = z.infer<typeof loginSchema>;

// Register schema
export const registerSchema = z.object({
  email: z
    .string()
    .min(1, 'auth.emailRequired')
    .email('auth.emailInvalid'),
  password: z
    .string()
    .min(1, 'auth.passwordRequired')
    .min(8, 'auth.passwordMinLength')
    .regex(/[A-Z]/, 'Password must contain at least one uppercase letter')
    .regex(/[a-z]/, 'Password must contain at least one lowercase letter')
    .regex(/[0-9]/, 'Password must contain at least one number'),
  confirmPassword: z
    .string()
    .min(1, 'Confirm password is required'),
  firstName: z.string().optional(),
  lastName: z.string().optional(),
}).refine((data) => data.password === data.confirmPassword, {
  message: 'auth.passwordMismatch',
  path: ['confirmPassword'],
});

export type RegisterFormData = z.infer<typeof registerSchema>;
