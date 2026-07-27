import { NextRequest, NextResponse } from 'next/server'
import { prisma } from '@/lib/prisma'
import { getCurrentUser } from '@/lib/session'
import { handleError, getJsonBody } from '@/lib/api-utils'
import { feedbackSchema } from '@/lib/schemas'

/**
 * GET /api/feedback
 * List user's feedback submissions
 */
export async function GET(request: NextRequest) {
  try {
    const session = await getCurrentUser()
    if (!session) {
      return NextResponse.json({ error: 'Unauthorized' }, { status: 401 })
    }

    const feedback = await prisma.feedback.findMany({
      where: { userId: session.userId },
      orderBy: { createdAt: 'desc' },
      include: {
        replies: {
          orderBy: { createdAt: 'asc' },
          select: {
            id: true,
            message: true,
            createdAt: true,
            isAdmin: true,
          }
        }
      }
    })

    return NextResponse.json({ feedback })
  } catch (error) {
    return handleError(error)
  }
}

/**
 * POST /api/feedback
 * Create new feedback
 */
export async function POST(request: NextRequest) {
  try {
    const session = await getCurrentUser()
    if (!session) {
      return NextResponse.json({ error: 'Unauthorized' }, { status: 401 })
    }

    const body = await getJsonBody(request)
    const validated = feedbackSchema.parse(body)

    const feedback = await prisma.feedback.create({
      data: {
        userId: session.userId,
        subject: validated.category,
        message: validated.message,
        status: 'open',
      }
    })

    return NextResponse.json({ feedback }, { status: 201 })
  } catch (error) {
    return handleError(error)
  }
}
