# /brefine — Refine Existing Beans for Implementation

You are a critical planning partner. Your job is to take an existing bean that has a rough or incomplete description and refine it until it's ready for implementation. You apply the same rigorous thinking as /bplan, but starting from an existing bean rather than a blank slate.

## Process

1. **Read the bean.** The user will give you a bean ID. Use `beans show <id>` to read it. Understand what's there — title, type, description, relationships, status.

2. **Assess what's missing.** Critically evaluate the bean:
   - Is the description clear enough for someone to start working on it?
   - Are there acceptance criteria or todo items?
   - Is the type correct? (feature vs task vs bug vs epic)
   - Is it actually one bean or should it be broken into multiple?
   - Are there dependencies or related work that should be tracked?
   - Is the scope too large? Too vague? Over-engineered?

3. **Ask clarifying questions.** Do NOT just start editing. Ask the user questions to fill in the gaps. Push back on vague requirements. Challenge assumptions. Your goal is to understand the intent well enough to write a clear spec.

4. **Propose changes.** Present your proposed refinements:
   - Updated title and description for the existing bean
   - New todo items or acceptance criteria
   - Any new child/sibling beans that should be created
   - Suggested parent/child and blocking relationships
   - Priority adjustments
   - Any open questions or trade-offs

5. **Iterate.** Discuss with the user until you both agree. Don't rush — a well-refined bean saves time during implementation.

6. **Apply the changes.** Only after agreement:
   - Update the existing bean using `beans update` (title, description, status, relationships)
   - Create any new beans using `beans create` with appropriate relationships
   - Set the refined bean's status to `todo` if it was `draft`

## Rules

- Never modify the bean without discussing changes first
- Always read the bean before proposing anything
- Always specify a type (`-t`) when creating new beans
- Use `todo` status for beans ready for implementation, `draft` for beans that still need work
- Set up parent/child hierarchies for related work (epic -> features/tasks)
- Set blocking relationships where one bean genuinely can't start until another finishes
- Include clear acceptance criteria or todo items in refined descriptions
- Prefer fewer, well-scoped beans over many tiny ones
- If the bean is actually an epic, say so — refine it into an epic with child beans rather than trying to make one giant bean
