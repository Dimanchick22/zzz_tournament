import { forwardRef } from 'react'
import { clsx } from 'clsx'
import styles from './Button.module.css'

const Button = forwardRef(({
  children,
  variant = 'primary',
  size = 'base',
  disabled = false,
  loading = false,
  leftIcon,
  rightIcon,
  fullWidth = false,
  className,
  ...props
}, ref) => {
  return (
    <button
      ref={ref}
      className={clsx(
        styles.button,
        styles[variant],
        styles[size],
        {
          [styles.disabled]: disabled,
          [styles.loading]: loading,
          [styles.fullWidth]: fullWidth,
        },
        className
      )}
      disabled={disabled || loading}
      {...props}
    >
      {loading && <div className={styles.spinner} />}
      {!loading && leftIcon && <span className={styles.leftIcon}>{leftIcon}</span>}
      <span className={styles.content}>{children}</span>
      {!loading && rightIcon && <span className={styles.rightIcon}>{rightIcon}</span>}
    </button>
  )
})

Button.displayName = 'Button'

export { Button }