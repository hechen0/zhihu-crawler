from time import sleep
import subprocess

def main():
  files_to_commit = subprocess.check_output("git ls-files . --others --exclude-standard", shell=True).strip().split("\n")

  total = len(files_to_commit)
  for i in range(total):
    j = i + 1
    if j % 3 == 0:
      cmd = "git add %s %s %s; git commit -m 'left %s'; git push;" % (files_to_commit[i - 2], files_to_commit[i-1], files_to_commit[i], total - i)
      subprocess.call(cmd, shell=True)
      sleep(.5)

if __name__ == "__main__":
	main()
